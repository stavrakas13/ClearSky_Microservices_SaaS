require('dotenv').config();
const amqp = require('amqplib');
const mysql = require('mysql2/promise');
const XLSX = require('xlsx');

/* const {
  MYSQL_URI,
  RABBITMQ_URI,
  RABBITMQ_EXCHANGE,
  RABBITMQ_ROUTING_KEY,
  RABBITMQ_SEND_AVAIL_KEY,
  RABBITMQ_GET_GRADES_KEY
} = process.env;

if (!MYSQL_URI || !RABBITMQ_URI || !RABBITMQ_EXCHANGE || !RABBITMQ_ROUTING_KEY) {
  console.error('❌  Missing required environment variables');
  process.exit(1);
}
 */

const {
  MYSQL_URI,
  RABBITMQ_URI,
  RABBITMQ_EXCHANGE,
  RABBITMQ_ROUTING_KEY,
  RABBITMQ_SEND_AVAIL_KEY,
  RABBITMQ_GET_GRADES_KEY
} = process.env;

const missingVars = [];

if (!MYSQL_URI) missingVars.push('MYSQL_URI');
if (!RABBITMQ_URI) missingVars.push('RABBITMQ_URI');
if (!RABBITMQ_EXCHANGE) missingVars.push('RABBITMQ_EXCHANGE');
if (!RABBITMQ_ROUTING_KEY) missingVars.push('RABBITMQ_ROUTING_KEY');

if (missingVars.length > 0) {
  console.error(`❌ Missing required environment variables: ${missingVars.join(', ')}`);
  process.exit(1);
}


(async () => {
  let connection;
  try {
    connection = await mysql.createConnection(MYSQL_URI);
    console.log('✅  Connected to MySQL');
  } catch (err) {
    console.error('❌ Failed to connect to MySQL:', err.message);
    process.exit(1);
  }

  let conn, channel;
  try {
    conn = await amqp.connect(RABBITMQ_URI);
    channel = await conn.createChannel();
    await channel.assertExchange(RABBITMQ_EXCHANGE, 'direct', { durable: true });
    console.log('✅  Connected to RabbitMQ and exchange set');
  } catch (err) {
    console.error('❌ RabbitMQ connection/setup failed:', err.message);
    process.exit(1);
  }

  const makeReply = msg => payload => {
    const { replyTo, correlationId } = msg.properties;
    if (!replyTo) return;
    try {
      channel.publish('', replyTo, Buffer.from(JSON.stringify(payload)), {
        contentType: 'application/json',
        correlationId
      });
      console.log(`📤  Reply sent to ${replyTo} (corrId: ${correlationId})`);
    } catch (err) {
      console.error('❌ Failed to send reply:', err.message);
    }
  };

   const q = 'postgrades.final';
  await channel.assertQueue(q, { durable: true, exclusive: false, autoDelete: false });
  await channel.bindQueue(q, RABBITMQ_EXCHANGE, RABBITMQ_ROUTING_KEY);
  channel.prefetch(10);
  console.log(`🚀  Listening for grades on "${RABBITMQ_ROUTING_KEY}"`);

  channel.consume(q, async msg => {
    if (!msg) return;
    console.log('📩  Received grade message');
    const reply = makeReply(msg);

    const ct = (msg.properties.contentType || '').toLowerCase().trim();
    const buffer = (ct.includes('spreadsheet') || ct === 'application/octet-stream')
      ? msg.content
      : Buffer.from(msg.content.toString(), 'base64');

    try {
      const wb = XLSX.read(buffer, { type: 'buffer' });
      const rows = XLSX.utils.sheet_to_json(wb.Sheets[wb.SheetNames[0]], {
        header: 1,
        raw: false
      });

      if (rows.length < 4) throw new Error('Template too short');
      console.log(`📊  Parsed XLSX with ${rows.length} rows`);

      const weightRow = rows[1], headerRow = rows[2], dataRows = rows.slice(3);

      const map = {
        'Αριθμός Μητρώου': 'AM',
        'Ονοματεπώνυμο': 'name',
        'Ακαδημαϊκό E-mail': 'email',
        'Περίοδος δήλωσης': 'declarationPeriod',
        'Τμήμα Τάξης': 'classTitle',
        'Κλίμακα βαθμολόγησης': 'gradingScale',
        'Βαθμολογία': 'grade'
      };

      let totalInserted = 0;

      for (const row of dataRows) {
        const d = {};
        headerRow.forEach((t, i) => {
          const k = map[t?.trim()];
          if (k && row[i] != null && row[i] !== '')
            d[k] = k === 'grade' ? parseFloat(row[i]) : row[i].toString().trim();
        });

        for (let q = 1; q <= 10; q++) {
          const idx = 8 + (q - 1);
          const score = parseFloat(row[idx]);
          const weight = parseFloat(weightRow[idx]);
          d[`Q${q}`] = (!isNaN(score) && !isNaN(weight)) ? score * weight : null;
        }

        const now = new Date();

        // Upsert grading
        const gradingSql = `
          INSERT INTO grading (
            AM, name, email, declarationPeriod, classTitle,
            gradingScale, grade,
            Q1, Q2, Q3, Q4, Q5, Q6, Q7, Q8, Q9, Q10
          ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
          ON DUPLICATE KEY UPDATE
            name = VALUES(name),
            email = VALUES(email),
            gradingScale = VALUES(gradingScale),
            grade = VALUES(grade),
            Q1 = VALUES(Q1), Q2 = VALUES(Q2), Q3 = VALUES(Q3), Q4 = VALUES(Q4), Q5 = VALUES(Q5),
            Q6 = VALUES(Q6), Q7 = VALUES(Q7), Q8 = VALUES(Q8), Q9 = VALUES(Q9), Q10 = VALUES(Q10)
        `;

        await connection.execute(gradingSql, [
          d.AM, d.name, d.email, d.declarationPeriod, d.classTitle,
          d.gradingScale, d.grade,
          d.Q1, d.Q2, d.Q3, d.Q4, d.Q5, d.Q6, d.Q7, d.Q8, d.Q9, d.Q10
        ]);

          if (totalInserted==0) {
          // Submission log handling (without AM)
          const [existingLog] = await connection.execute(
            `SELECT initialSubmissionDate, finalSubmissionDate 
            FROM submission_log 
            WHERE declarationPeriod = ? AND classTitle = ?`,
            [d.declarationPeriod, d.classTitle]
          );
          if (existingLog.length === 0) {
            await connection.execute(
              `INSERT INTO submission_log (declarationPeriod, classTitle, initialSubmissionDate)
              VALUES (?, ?, ?)`,
              [d.declarationPeriod, d.classTitle, now]
            );
          } else {
            await connection.execute(
              `UPDATE submission_log SET finalSubmissionDate = ? 
              WHERE declarationPeriod = ? AND classTitle = ?`,
              [now, d.declarationPeriod, d.classTitle]
            );
          }
        }
        totalInserted++;

      }
      

      console.log(`✅  Processed ${totalInserted} grades`);
      reply({ status: 'ok', message: `Processed ${totalInserted} grades` });
      channel.ack(msg);

    } catch (err) {
      console.error('❌ Error processing grades:', err.message);
      reply({ status: 'error', message: err.message });
      channel.nack(msg, false, false);
    }
  }, { noAck: false });
  {
    const q2 = 'get.submission.logs';
    await channel.assertQueue(q2, { durable: false, exclusive: false, autoDelete: false });
    await channel.bindQueue(q2, RABBITMQ_EXCHANGE, RABBITMQ_SEND_AVAIL_KEY);
    console.log(`📥  Listening for submission-log requests on "${RABBITMQ_SEND_AVAIL_KEY}"`);

    channel.consume(q2, async msg => {
      if (!msg) return;
      console.log('📩  Received submission-log request');
      const reply = makeReply(msg);

      try {
        const [rows] = await connection.execute(`SELECT * FROM submission_log`);
        reply({ status: 'ok', data: rows });
        channel.ack(msg);
      } catch (err) {
        console.error('❌ Error fetching submission logs:', err.message);
        reply({ status: 'error', message: err.message });
        channel.nack(msg, false, false);
      }
    }, { noAck: false });
  }

// -- histogram helper with dynamic upper‐bound on bins --
async function fetchHistogram(field, connection, { classTitle, declarationPeriod }) {
  // round the float into integer bins
  const sql = `
    SELECT
      ROUND(\`${field}\`) AS value,
      COUNT(*)           AS count
    FROM grading
    WHERE classTitle       = ?
      AND declarationPeriod = ?
    GROUP BY ROUND(\`${field}\`)
    ORDER BY ROUND(\`${field}\`)
  `;
  const [rows] = await connection.execute(sql, [ classTitle, declarationPeriod ]);

  // if it's the overall 'grade', force 0–10; otherwise use max rounded value
  const maxBin = field === 'grade'
    ? 10
    : rows.reduce((max, r) => Math.max(max, +r.value), 0);

  // build bins 0 through maxBin
  const categories = Array.from({ length: maxBin + 1 }, (_, i) => i);

  // map counts into those bins
  const data = categories.map(i => {
    const found = rows.find(r => +r.value === i);
    return found ? +found.count : 0;
  });

  return { categories, data };
}


// -- RabbitMQ consumer that calls fetchHistogram for each dimension --
{
  const q3 = 'get.grades';
  await channel.assertQueue(q3, { durable: false, exclusive: false, autoDelete: false });
  await channel.bindQueue(q3, RABBITMQ_EXCHANGE, RABBITMQ_GET_GRADES_KEY);
  console.log(`📥  Listening for grade-fetch requests on "${RABBITMQ_GET_GRADES_KEY}"`);

  channel.consume(q3, async msg => {
    if (!msg) return;
    console.log('📩  Received get.grades request');
    const reply = makeReply(msg);

    let params;
    try {
      params = JSON.parse(msg.content.toString());
    } catch (err) {
      console.error('❌ Invalid JSON payload:', err.message);
      reply({ status: 'error', message: 'Invalid JSON payload' });
      return channel.nack(msg, false, false);
    }

    const { declarationPeriod, classTitle } = params;
    if (!declarationPeriod || !classTitle) {
      const missing = ['declarationPeriod', 'classTitle']
        .filter(k => !params[k]).join(', ');
      reply({ status: 'error', message: `Missing fields: ${missing}` });
      return channel.ack(msg);
    }

    try {
      // Build histograms for total grade + Q1–Q10
      const dims = ['grade', 'Q1', 'Q2', 'Q3', 'Q4', 'Q5', 'Q6', 'Q7', 'Q8', 'Q9', 'Q10'];
      const result = {};
      for (let dim of dims) {
        result[dim] = await fetchHistogram(dim, connection, params);
      }

      reply({ status: 'ok', data: result });
      channel.ack(msg);

    } catch (err) {
      console.error('❌ Error fetching grade histograms:', err.message);
      reply({ status: 'error', message: err.message });
      channel.nack(msg, false, false);
    }
  }, { noAck: false });
}
})();

require('dotenv').config();
const amqp = require('amqplib');
const mysql = require('mysql2/promise');
const XLSX = require('xlsx');

const {
  MYSQL_URI,
  RABBITMQ_URI,
  RABBITMQ_EXCHANGE,
  RABBITMQ_ROUTING_KEY,
  RABBITMQ_GET_GRADES_KEY
} = process.env;

const missingVars = [];
if (!MYSQL_URI) missingVars.push("MYSQL_URI");
if (!RABBITMQ_URI) missingVars.push("RABBITMQ_URI");
if (!RABBITMQ_EXCHANGE) missingVars.push("RABBITMQ_EXCHANGE");
if (!RABBITMQ_ROUTING_KEY) missingVars.push("RABBITMQ_ROUTING_KEY");
if (!RABBITMQ_GET_GRADES_KEY) missingVars.push("RABBITMQ_GET_GRADES_KEY");

if (missingVars.length > 0) {
  console.error(`[${new Date().toISOString()}] ❌ Missing required environment variables: ${missingVars.join(", ")}`);
  process.exit(1);
}

const log = (...args) => console.log(`[${new Date().toISOString()}]`, ...args);

(async () => {
  // Connect to MySQL
  let connection;
  try {
    connection = await mysql.createConnection(MYSQL_URI);
    log('✅ Connected to MySQL');
  } catch (err) {
    log('❌ Failed to connect to MySQL:', err.message);
    process.exit(1);
  }

  // Connect to RabbitMQ & exchange
  let conn, channel;
  try {
    conn = await amqp.connect(RABBITMQ_URI);
    channel = await conn.createChannel();
    await channel.assertExchange(RABBITMQ_EXCHANGE, 'direct', { durable: true });
    log('✅ Connected to RabbitMQ and exchange set');
  } catch (err) {
    log('❌ RabbitMQ connection/setup failed:', err.message);
    process.exit(1);
  }

  // Helper to reply on the same correlationId & replyTo
  const makeReply = msg => payload => {
    const { replyTo, correlationId } = msg.properties;
    if (!replyTo) {
      log('⚠️ No replyTo queue specified; skipping reply');
      return;
    }
    try {
      channel.publish('', replyTo, Buffer.from(JSON.stringify(payload)), {
        contentType: 'application/json',
        correlationId
      });
      log(`📤 Reply sent (corrId=${correlationId}) to ${replyTo}`);
    } catch (err) {
      log('❌ Failed to send reply:', err.message);
    }
  };

  // ───────────────────────────────────────────────────────────────────────────
  // 1️⃣ Grade Import via XLSX
  const importQueue = 'postgrades.final';
  await channel.assertQueue(importQueue, { durable: true, exclusive: false, autoDelete: false });
  await channel.bindQueue(importQueue, RABBITMQ_EXCHANGE, RABBITMQ_ROUTING_KEY);
  channel.prefetch(10);
  log(`🚀 Listening for XLSX uploads on "${RABBITMQ_ROUTING_KEY}"`);

  channel.consume(importQueue, async msg => {
    if (!msg) return;
    log('📩 Received XLSX grade message');
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
      log(`📊 Parsed XLSX with ${rows.length} rows`);

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

        const gradingSql = `
          INSERT INTO grading (
            AM, name, email, declarationPeriod, classTitle,
            gradingScale, grade,
            Q1, Q2, Q3, Q4, Q5, Q6, Q7, Q8, Q9, Q10, grading_status
          ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
          ON DUPLICATE KEY UPDATE
            name = VALUES(name),
            email = VALUES(email),
            gradingScale = VALUES(gradingScale),
            grade = VALUES(grade),
            Q1 = VALUES(Q1), Q2 = VALUES(Q2), Q3 = VALUES(Q3), Q4 = VALUES(Q4), Q5 = VALUES(Q5),
            Q6 = VALUES(Q6), Q7 = VALUES(Q7), Q8 = VALUES(Q8), Q9 = VALUES(Q9), Q10 = VALUES(Q10),
            grading_status = 1
        `;

        await connection.execute(gradingSql, [
          d.AM, d.name, d.email, d.declarationPeriod, d.classTitle,
          d.gradingScale, d.grade,
          d.Q1, d.Q2, d.Q3, d.Q4, d.Q5, d.Q6, d.Q7, d.Q8, d.Q9, d.Q10
        ]);

        log(`✅ Upserted grade for AM=${d.AM}, class=${d.classTitle}`);
        totalInserted++;
      }

      log(`🎉 Imported total of ${totalInserted} grades`);
      reply({ status: 'ok', message: `Processed ${totalInserted} grades` });
      channel.ack(msg);
    } catch (err) {
      log('❌ Error importing grades:', err.message);
      reply({ status: 'error', message: err.message });
      channel.nack(msg, false, false);
    }
  }, { noAck: false });

  // ───────────────────────────────────────────────────────────────────────────
  // 2️⃣ Query Grades by AM
  const getGradesQueue = 'grades.get.byAM.q';
  await channel.assertQueue(getGradesQueue, { durable: true, exclusive: false, autoDelete: false });
  await channel.bindQueue(getGradesQueue, RABBITMQ_EXCHANGE, RABBITMQ_GET_GRADES_KEY);
  log(`🎓 Listening for grade queries on "${RABBITMQ_GET_GRADES_KEY}"`);

  channel.consume(getGradesQueue, async (msg) => {
    if (!msg) return;
    log('📩 Received AM query message');

    // ─── RAW PAYLOAD ─────────────────────────────────────────────────────
    const raw = msg.content.toString();
    log(`📥 [getGrades] raw message: ${raw}`);

    // ─── PARSED BODY ─────────────────────────────────────────────────────
    let body;
    try {
      body = JSON.parse(raw);
      log('📥 [getGrades] parsed body:', body);
    } catch (e) {
      log('❌ [getGrades] JSON parse error:', e.message);
      const reply = makeReply(msg);
      reply({ status: 'error', error: 'Invalid JSON' });
      return channel.nack(msg, false, false);
    }

    const reply = makeReply(msg);

    try {
      // Accept either body.AM or body.student_id depending on what your Go service sends
      const am = ((body.AM || body.student_id) || '').trim();

      if (!am) {
        log('⚠️ AM is missing from request');
        throw new Error('Missing AM in request');
      }

      log(`🔍 Looking up grades for AM=${am}`);
      const [rows] = await connection.execute(
        `SELECT declarationPeriod, classTitle, grading_status, grade
           FROM grading
          WHERE AM = ?`,
        [am]
      );

      log(`📤 Found ${rows.length} grade(s) for AM ${am}`);
      reply({ status: 'ok', data: rows });
      channel.ack(msg);
    } catch (err) {
      log('❌ Failed to handle AM request:', err.message);
      reply({ status: 'error', error: err.message });
      channel.nack(msg, false, false);
    }
  }, { noAck: false });

})();

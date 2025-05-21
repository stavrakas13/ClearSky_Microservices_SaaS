/**
 * grades_consumer.js
 *
 * Listens on topic "postgrades.init", ingests an .xlsx that may arrive as
 *  • ✉️  raw binary  (contentType = application/vnd.openxmlformats-officedocument.spreadsheetml.sheet)
 *  • ✉️  base-64 string, contentType omitted  (your CLI test)
 *  • ✉️  base-64 string, contentType = text/plain
 * Inserts the parsed rows into MongoDB and (optionally) replies RPC-style.
 */
require('dotenv').config();
const amqp     = require('amqplib');
const mongoose = require('mongoose');
const XLSX     = require('xlsx');

// ─────────────────────── ENV ───────────────────────
const {
  MONGO_URI,
  RABBITMQ_URI,
  RABBITMQ_EXCHANGE,   // "clearSky.event"
  RABBITMQ_ROUTING_KEY // "postgrades.init"
} = process.env;

if (!MONGO_URI || !RABBITMQ_URI || !RABBITMQ_EXCHANGE || !RABBITMQ_ROUTING_KEY) {
  console.error('❌  Missing .env variables');
  process.exit(1);
}

// ────────────────── Mongoose model ─────────────────
const GradeSchema = new mongoose.Schema({
  AM:    { type: String, required: true },
  name:  { type: String, required: true },
  email: { type: String, required: true },
  declarationPeriod:{ type: String, required: true },
  classTitle:       { type: String, required: true },
  gradingScale:     { type: String, required: true },
  grade:            { type: Number, required: true },
  Q1:  { type: Number, min:0, max:1000, default:null },
  Q2:  { type: Number, min:0, max:1000, default:null },
  Q3:  { type: Number, min:0, max:1000, default:null },
  Q4:  { type: Number, min:0, max:1000, default:null },
  Q5:  { type: Number, min:0, max:1000, default:null },
  Q6:  { type: Number, min:0, max:1000, default:null },
  Q7:  { type: Number, min:0, max:1000, default:null },
  Q8:  { type: Number, min:0, max:1000, default:null },
  Q9:  { type: Number, min:0, max:1000, default:null },
  Q10: { type: Number, min:0, max:1000, default:null }
});
const Grade = mongoose.model('Grade', GradeSchema);

// ──────────────────── Main worker ──────────────────
(async () => {
  await mongoose.connect(MONGO_URI);
  console.log('✅  MongoDB connected');

  const conn    = await amqp.connect(RABBITMQ_URI);
  const channel = await conn.createChannel();

  await channel.assertExchange(RABBITMQ_EXCHANGE, 'topic', { durable: true });
  const q = await channel.assertQueue('', { exclusive: true });
  await channel.bindQueue(q.queue, RABBITMQ_EXCHANGE, RABBITMQ_ROUTING_KEY);
  channel.prefetch(10);

  console.log(`🚀  Waiting on ${RABBITMQ_EXCHANGE} → "${RABBITMQ_ROUTING_KEY}"`);

  channel.consume(q.queue, async (msg) => {
    if (!msg) return;

    // ---------- 1) Figure out what we actually received ----------
    const ct = (msg.properties.contentType || '').toLowerCase().trim();
    let buffer;

    if (ct === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' ||
        ct === 'application/octet-stream') {
      buffer = msg.content;                          // raw XLSX
    } else {
      // either text/plain OR header completely missing → treat as base-64 string
      buffer = Buffer.from(msg.content.toString(), 'base64');
    }

    // ---------- helper: optional RPC reply ----------
    const reply = (payload) => {
      const { replyTo, correlationId } = msg.properties;
      if (!replyTo) return;                          // not an RPC call
      channel.publish(
        '', replyTo,
        Buffer.from(JSON.stringify(payload)),
        { contentType: 'application/json', correlationId }
      );
    };

    try {
      // ---------- 2) Parse workbook ----------
      const wb   = XLSX.read(buffer, { type: 'buffer' });
      const rows = XLSX.utils.sheet_to_json(wb.Sheets[wb.SheetNames[0]], {
        header: 1, raw: false
      });
      if (rows.length < 4) throw new Error('Template too short');

      const weightRow = rows[1];
      const headerRow = rows[2];
      const dataRows  = rows.slice(3);

      const map = {
        'Αριθμός Μητρώου':   'AM',
        'Ονοματεπώνυμο':     'name',
        'Ακαδημαϊκό E-mail': 'email',
        'Περίοδος δήλωσης':  'declarationPeriod',
        'Τμήμα Τάξης':       'classTitle',
        'Κλίμακα βαθμολόγησης': 'gradingScale',
        'Βαθμολογία':        'grade'
      };

      const docs = dataRows.map(row => {
        const d = {};
        headerRow.forEach((title, i) => {
          const k = map[title?.trim()];
          if (!k) return;
          if (row[i] != null && row[i] !== '')
            d[k] = k === 'grade' ? parseFloat(row[i]) : row[i].toString().trim();
        });
        for (let q = 1; q <= 10; q++) {
          const idx = 8 + (q - 1);
          const score  = parseFloat(row[idx]);
          const weight = parseFloat(weightRow[idx]);
          d[`Q${q}`] = (!isNaN(score) && !isNaN(weight)) ? score * weight : null;
        }
        return d;
      });

      const res = await Grade.insertMany(docs, { ordered: false });
      console.log(`✅  Inserted ${res.length} docs`);

      reply({ status: 'ok', message: `Inserted ${res.length} records` });
      channel.ack(msg);

    } catch (err) {
      console.error('❌', err.message);
      reply({ status: 'error', message: 'Failed to process file', error: err.message });
      channel.nack(msg, false, false); // discard
    }
  }, { noAck: false });
})();

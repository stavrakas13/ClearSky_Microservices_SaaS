require('dotenv').config();
const amqp = require('amqplib');
const mongoose = require('mongoose');
const XLSX = require('xlsx');

// Pull straight from your .env
const {
  MONGO_URI,
  RABBITMQ_URI,
  RABBITMQ_EXCHANGE,
  RABBITMQ_ROUTING_KEY
} = process.env;

// (Optional) sanity‐check:
if (!MONGO_URI || !RABBITMQ_URI || !RABBITMQ_EXCHANGE || !RABBITMQ_ROUTING_KEY) {
  console.error('❌ Missing one of MONGO_URI, RABBITMQ_URI, RABBITMQ_EXCHANGE or RABBITMQ_ROUTING_KEY in .env');
  process.exit(1);
}

// 1. Define schema with Q1–Q10
const GradeSchema = new mongoose.Schema({
  AM:               { type: String,  required: true },
  name:             { type: String,  required: true },
  email:            { type: String,  required: true },
  declarationPeriod:{ type: String,  required: true },
  classTitle:       { type: String,  required: true },
  gradingScale:     { type: String,  required: true },
  grade:            { type: Number,  required: true },
  Q1: { type: Number, min: 0, max:1000, default: null },
  Q2: { type: Number, min: 0, max:1000, default: null },
  Q3: { type: Number, min: 0, max:1000, default: null },
  Q4: { type: Number, min: 0, max:1000, default: null },
  Q5: { type: Number, min: 0, max:1000, default: null },
  Q6: { type: Number, min: 0, max:1000, default: null },
  Q7: { type: Number, min: 0, max:1000, default: null },
  Q8: { type: Number, min: 0, max:1000, default: null },
  Q9: { type: Number, min: 0, max:1000, default: null },
  Q10:{ type: Number, min: 0, max:1000, default: null }
});
const Grade = mongoose.model('Grade', GradeSchema);

async function connectMongo() {
  await mongoose.connect(MONGO_URI, { useNewUrlParser: true, useUnifiedTopology: true });
  console.log('✅ Connected to MongoDB');
}

async function startConsumer() {
  const conn = await amqp.connect(RABBITMQ_URI);
  const channel = await conn.createChannel();
  await channel.assertExchange(RABBITMQ_EXCHANGE, 'topic', { durable: true });
  const q = await channel.assertQueue('', { exclusive: true });
  await channel.bindQueue(q.queue, RABBITMQ_EXCHANGE, RABBITMQ_ROUTING_KEY);
  console.log(`🚀 Waiting for messages on ${RABBITMQ_EXCHANGE} with routing key ${RABBITMQ_ROUTING_KEY}`);

  channel.consume(q.queue, async (msg) => {
    if (!msg) return;
    try {
      const ct = msg.properties.contentType;

      if (ct === 'text/plain') {
        // Plain‐text path
        const text = msg.content.toString('utf8');
        console.log('📥 Received plain text:', text);
        channel.ack(msg);
        return;
      }

      // Excel/Base64 path
      console.log('📥 Received Excel/Base64 payload');
      const base64  = msg.content.toString();
      const buffer  = Buffer.from(base64, 'base64');
      const workbook= XLSX.read(buffer, { type: 'buffer' });
      const sheet   = workbook.Sheets[workbook.SheetNames[0]];
      const rows    = XLSX.utils.sheet_to_json(sheet, { header: 1, raw: false });

      // Row 2 (index 1) has weights columns I–R → indexes 8–17
      const weightRow = rows[1] || [];

      // Row 3 (index 2) is the header names
      const headerRow = rows[2] || [];

      // Data rows start at row 4 → index 3+
      const dataRows  = rows.slice(3);

      // Fixed-field mapping
      const mapping = {
        'Αριθμός Μητρώου':   'AM',
        'Ονοματεπώνυμο':     'name',
        'Ακαδημαϊκό E-mail': 'email',
        'Περίοδος δήλωσης':  'declarationPeriod',
        'Τμήμα Τάξης':       'classTitle',
        'Κλίμακα βαθμολόγησης': 'gradingScale',
        'Βαθμολογία':         'grade'
      };

      const docs = dataRows.map(row => {
        const doc = {};

        // 1) Map fixed fields
        headerRow.forEach((col, i) => {
          const key = mapping[col && col.trim()];
          if (key && row[i] != null && row[i] !== '') {
            // cast grade to Number
            doc[key] = (key === 'grade') ? parseFloat(row[i]) : row[i].toString().trim();
          }
        });

        // 2) Compute Q1–Q10 from cols I (8) → R (17)
        for (let q = 1; q <= 10; q++) {
          const idx     = 8 + (q - 1);
          const rawVal  = parseFloat(row[idx]);
          const weight  = parseFloat(weightRow[idx]);
          if (!isNaN(rawVal) && !isNaN(weight)) {
            doc[`Q${q}`] = rawVal * weight;
          } else {
            doc[`Q${q}`] = null;
          }
        }

        return doc;
      });

      // 3) Bulk insert
      const result = await Grade.insertMany(docs);
      console.log(`✅ Inserted ${result.length} records into MongoDB`);
      channel.ack(msg);

    } catch (err) {
      console.error('❌ Error processing message', err);
      channel.nack(msg, false, false);
    }
  }, { noAck: false });
}

(async () => {
  try {
    await connectMongo();
    await startConsumer();
  } catch (err) {
    console.error('❌ Failed to start service', err);
    process.exit(1);
  }
})();

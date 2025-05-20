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

const GradeSchema = new mongoose.Schema({
  AM: { type: String, required: true },
  name: { type: String, required: true },
  email: { type: String, required: true },
  declarationPeriod: { type: String, required: true },
  classTitle: { type: String, required: true },
  gradingScale: { type: String, required: true },
  grade: { type: Number, required: true },
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
    if (msg) {
      try {
        console.log('📥 Received message');
        const base64 = msg.content.toString();
        const buffer = Buffer.from(base64, 'base64');
        const workbook = XLSX.read(buffer, { type: 'buffer' });
        const sheet = workbook.Sheets[workbook.SheetNames[0]];

        const rows = XLSX.utils.sheet_to_json(sheet, { header: 1, raw: false });
        const headerRow = rows[2];
        const dataRows = rows.slice(3);

        const mapping = {
          'Αριθμός Μητρώου': 'AM',
          'Ονοματεπώνυμο': 'name',
          'Ακαδημαϊκό E-mail': 'email',
          'Περίοδος δήλωσης': 'declarationPeriod',
          'Τμήμα Τάξης': 'classTitle',
          'Κλίμακα βαθμολόγησης': 'gradingScale',
          'Βαθμολογία': 'grade'
        };

        const docs = dataRows.map(row => {
          const doc = {};
          headerRow.forEach((col, i) => {
            const key = mapping[col && col.trim()];
            if (key) doc[key] = row[i];
          });
          return doc;
        });

        await Grade.insertMany(docs);
        console.log(`✅ Inserted ${docs.length} records into MongoDB`);
        channel.ack(msg);
      } catch (err) {
        console.error('❌ Error processing message', err);
        channel.nack(msg, false, false);
      }
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

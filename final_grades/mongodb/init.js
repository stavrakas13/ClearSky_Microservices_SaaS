// init_credits.js
require('dotenv').config();
const { MongoClient } = require('mongodb');

async function main() {
  const uri    = process.env.MONGO_URI;
  const client = new MongoClient(uri, {
    useNewUrlParser:    true,
    useUnifiedTopology: true
  });

  try {
    await client.connect();
    const db = client.db('credits');

    // 1) Drop existing collection if any
    const exists = await db.listCollections({ name: 'credits' }).toArray();
    if (exists.length) {
      await db.collection('credits').drop();
      console.log('✓ Dropped existing collection "credits"');
    }

    // 2) Define the JSON-Schema validator
    const validator = {
      $jsonSchema: {
        bsonType: 'object',
        required: ['cred', 'name'],
        properties: {
          cred: {
            bsonType: 'int',
            description: 'How many credits does this organization have'
          },
          name: {
            bsonType: 'string',
            description: 'Organization encoded name'
          }
        }
      }
    };

    // 3) Create with validator
    await db.createCollection('credits', { validator });
    console.log('✓ Created "credits" collection with schema validation');

    // 4) Seed initial data
    const seedDocs = [
      { name: 'NTUA', cred: 50 },
      { name: 'EKPA', cred: 50 }
    ];
    const res = await db.collection('credits').insertMany(seedDocs);
    console.log(`✓ Inserted ${res.insertedCount} seed documents`);

  } catch (err) {
    console.error('❌ Error initializing credits:', err);
  } finally {
    await client.close();
  }
}

main();

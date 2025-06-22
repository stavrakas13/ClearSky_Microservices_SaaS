// create_grading.js
require('dotenv').config();
const { MongoClient } = require('mongodb');

async function main() {
  // αν δεν υπάρχει MONGO_URI στο .env, πέφτει στο localhost
  const uri = process.env.MONGO_URI || 'mongodb://localhost:27017';
  const client = new MongoClient(uri, { useUnifiedTopology: true });

  try {
    await client.connect();
    console.log('✅ Connected to MongoDB');

    const db = client.db('init_grades');
    const colName = 'grading';

    // 1) Αν υπάρχει ήδη, drop it
    const exists = await db.listCollections({ name: colName }).toArray();
    if (exists.length) {
      await db.collection(colName).drop();
      console.log(`✓ Dropped existing collection '${colName}'`);
    }

    // 2) JSON‐Schema validator
    const validator = {
      $jsonSchema: {
        bsonType: 'object',
        required: [
          'AM','name','email','declarationPeriod',
          'classTitle','gradingScale','grade'
        ],
        properties: {
          AM: {
            bsonType: 'string',
            description: 'Αριθμός Μητρώου — must be a string and is required'
          },
          name: {
            bsonType: 'string',
            description: 'Ονοματεπώνυμο — must be a string and is required'
          },
          email: {
            bsonType: 'string',
            pattern: '^.+@.+\\..+$',
            description: 'Academic email — must be a valid e-mail format'
          },
          declarationPeriod: {
            bsonType: 'string',
            description: 'Περίοδος δήλωσης — must be a string'
          },
          classTitle: {
            bsonType: 'string',
            description: 'Τμήμα Τάξης — must be a string'
          },
          gradingScale: {
            bsonType: 'string',
            description: 'Κλίμακα βαθμολόγησης — must be a string'
          },
          grade: {
            bsonType: 'number',
            minimum: 0,
            maximum: 10,
            description: 'Βαθμολογία — must be a number between 0 and 10'
          },
          // Q1–Q10: number 0–1000 ή null
          ...[...Array(10)].reduce((acc, _, i) => {
            const key = `Q${i+1}`;
            acc[key] = {
              bsonType: ['number','null'],
              minimum: 0,
              maximum: 1000,
              description: 'Optional numeric field 0–1000 or null'
            };
            return acc;
          }, {})
        }
      }
    };

    // 3) Δημιουργία collection
    await db.createCollection(colName, { validator });
    console.log(`✓ Created '${colName}' with schema validation`);

    // 4) Εμφάνιση λιστών για έλεγχο
    const cols = await db.listCollections().toArray();
    console.log('Collections in initial_grades:', cols.map(c=>c.name));

  } catch (err) {
    console.error('❌ Error in create_grading.js:', err);
  } finally {
    await client.close();
  }
}

main();

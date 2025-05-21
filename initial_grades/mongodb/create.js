// create.js
const { MongoClient } = require('mongodb');

async function main() {
  const uri = "mongodb://localhost:27017";
  const client = new MongoClient(uri);

  try {
    await client.connect();

    const db = client.db("grades");

    // 1. Drop existing collection if present
    const existing = await db.listCollections({ name: "grades" }).toArray();
    if (existing.length) {
      await db.collection("grades").drop();
      console.log("✓ Dropped existing collection 'grades'");
    }

    // 2. Define your JSON-Schema validator
    const validator = {
      $jsonSchema: {
        bsonType: "object",
        required: [
          "AM",
          "name",
          "email",
          "declarationPeriod",
          "classTitle",
          "gradingScale",
          "grade"
        ],
        properties: {
          AM: {
            bsonType: "string",
            description: "Αριθμός Μητρώου — must be a string and is required"
          },
          name: {
            bsonType: "string",
            description: "Ονοματεπώνυμο — must be a string and is required"
          },
          email: {
            bsonType: "string",
            pattern: "^.+@.+\\..+$",
            description: "Academic email — must be a valid e-mail format"
          },
          declarationPeriod: {
            bsonType: "string",
            description: "Περίοδος δήλωσης — must be a string"
          },
          classTitle: {
            bsonType: "string",
            description: "Τμήμα Τάξης — must be a string"
          },
          gradingScale: {
            bsonType: "string",
            description: "Κλίμακα βαθμολόγησης — must be a string"
          },
          grade: {
            bsonType: "number",
            minimum: 0,
            maximum: 10,
            description: "Βαθμολογία — must be a number between 0 and 10"
          },
          // Q1–Q10: number 0–1000 or null
          ...[...Array(10)].reduce((acc, _, i) => {
            const key = `Q${i + 1}`;
            acc[key] = {
              bsonType: ["number", "null"],
              minimum: 0,
              maximum: 1000,
              description: "Optional numeric field 0–1000 or null"
            };
            return acc;
          }, {})
        }
      }
    };

    // 3. Create collection with validator
    await db.createCollection("grades", { validator });
    console.log("✓ Created 'grades' collection with schema validation");

    // 4. Show databases
    const dbs = await client.db().admin().listDatabases();
    console.log("\nDatabases:");
    console.dir(dbs.databases, { depth: null });

    // 5. Show collections in 'grades'
    const cols = await db.listCollections().toArray();
    console.log("\nCollections in 'grades':");
    console.dir(cols, { depth: null });

    // 6. Show the validator
    const info = await db.listCollections({ name: "grades" }, { nameOnly: false }).toArray();
    console.log("\nValidator for 'grades':");
    console.dir(info[0].options.validator, { depth: null });

  } catch (err) {
    console.error(err);
  } finally {
    await client.close();
  }
}

main();

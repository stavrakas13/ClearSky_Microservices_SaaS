const { MongoClient } = require('mongodb');

async function main() {
  const uri = "mongodb://localhost:27017";
  const client = new MongoClient(uri);

  try {
    await client.connect();

    const db = client.db("credits");

    const existing = await db.listCollections({ name: "credits" }).toArray();
    if (existing.length) {
      await db.collection("credits").drop();
      console.log("✓ Dropped existing collection 'credits'");
    }

    const validator = {
      $jsonSchema: {
        bsonType: "object",
        required: ["cred", "name"],
        properties: {
          cred: {
            bsonType: "int",
            description: "How many credits does this organization have"
          },
          name: {
            bsonType: "string",
            description: "Organization encoded name"
          }
        }
      }
    };

    await db.createCollection("credits", { validator });
    console.log("✓ Created 'credits' collection with schema validation");

    const dbs = await client.db().admin().listDatabases();
    console.log("\nDatabases:");
    console.dir(dbs.databases, { depth: null });

    const cols = await db.listCollections().toArray();
    console.log("\nCollections in 'credits':");
    console.dir(cols, { depth: null });

    const info = await db.listCollections({ name: "credits" }, { nameOnly: false }).toArray();
    console.log("\nValidator for 'credits':");
    console.dir(info[0].options.validator, { depth: null });

  } catch (err) {
    console.error("❌ Error:", err);
  } finally {
    await client.close();
  }
}

main();

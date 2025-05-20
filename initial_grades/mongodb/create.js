
const db = connect("localhost:27017/grades");

// 2. Drop the collection if it already exists (optional)
if (db.getCollectionNames().includes("grades")) {
  db.grades.drop();
  print("✓ Dropped existing collection 'grades'");
}

// 3. Define the JSON-Schema validator
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
      }
    }
  }
};

// 4. Create the collection with validation
db.createCollection("grades", { validator });
print("✓ Created 'grades' collection with schema validation");

// 5. Optional: show the resulting setup
print("\nDatabases:");
printjson(db.adminCommand({ listDatabases: 1 }));

print("\nCollections in 'grades':");
printjson(db.getCollectionInfos({ name: "grades" }));

print("\nValidator for 'grades':");
printjson(db.getCollectionInfos({ name: "grades" })[0].options.validator);

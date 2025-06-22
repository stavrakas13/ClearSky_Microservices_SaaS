// test_grades.js
const mongoose = require('mongoose');
require('dotenv').config();

const {
  MONGODB_URI = 'mongodb://localhost:27017/init_grades'
} = process.env;

// re-use your Grade schema definition
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

(async () => {
  await mongoose.connect(MONGODB_URI, {
    useNewUrlParser: true,
    useUnifiedTopology: true
  });
  console.log('🔍 Connected, fetching 10 documents…');
  const docs = await Grade.find().limit(10).lean();
  console.log(docs);
  await mongoose.disconnect();
})();

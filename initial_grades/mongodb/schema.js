const mongoose = require("mongoose");

const GradeSchema = new mongoose.Schema({
  AM: { type: String, required: true },                     
  name: { type: String, required: true },                   
  email: { type: String, required: true },                 
  declarationPeriod: { type: String, required: true },     
  classTitle: { type: String, required: true },            
  gradingScale: { type: String, required: true },          
  grade: { type: Number, required: true },               
});


module.exports = mongoose.model("Grade", GradeSchema);

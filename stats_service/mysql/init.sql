-- Προτείνεται χρήση InnoDB για foreign keys
CREATE DATABASE IF NOT EXISTS stats CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE stats;

-- 1) Δημιουργία του course_declaration πίνακα
CREATE TABLE submission_log (
  declarationPeriod VARCHAR(50),
  classTitle VARCHAR(100),
  initialSubmissionDate DATETIME,
  finalSubmissionDate DATETIME,
  PRIMARY KEY (declarationPeriod, classTitle)
);


-- 2) Δημιουργία του grading πίνακα
CREATE TABLE grading (
  AM VARCHAR(20),
  name VARCHAR(100),
  email VARCHAR(100),
  declarationPeriod VARCHAR(50),
  classTitle VARCHAR(100),
  gradingScale VARCHAR(20),
  grade DECIMAL(4,2),
  Q1 INT, Q2 INT, Q3 INT, Q4 INT, Q5 INT,
  Q6 INT, Q7 INT, Q8 INT, Q9 INT, Q10 INT,
  PRIMARY KEY (AM, declarationPeriod, classTitle)
);


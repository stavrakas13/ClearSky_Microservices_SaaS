CREATE USER IF NOT EXISTS 'user'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON view_grades.* TO 'user'@'%';
FLUSH PRIVILEGES;

CREATE DATABASE IF NOT EXISTS view_grades CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE view_grades;

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
  grading_status TINYINT(1),
  PRIMARY KEY (AM, declarationPeriod, classTitle)
);

ALTER TABLE grading ADD COLUMN grading_status TINYINT(1) DEFAULT 0;



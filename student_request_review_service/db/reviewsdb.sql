-- for debugging purposes.
DROP TABLE IF EXISTS reviews;

-- reviews made by students added here.
-- combination of student_id & course_id & exam_period UNIQUE for each review. 
CREATE TABLE IF NOT EXISTS reviews (
  review_id SERIAL PRIMARY KEY,
  student_id VARCHAR(50) NOT NULL,
  course_id VARCHAR(50) NOT NULL,
  exam_period VARCHAR(50) NOT NULL,
  student_message TEXT NOT NULL,
  status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed')),
  instructor_reply_message TEXT,
  instructor_action VARCHAR(50) DEFAULT NULL CHECK (instructor_action IN ('Total accept', 'Partial accept', 'Reject')),
  review_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  reviewed_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS instructors (
  instructor_name VARCHAR(50) NOT NULL,
  course_id VARCHAR(50) NOT NULL
);

-- DEFAULT INSTRUCTOR

INSERT INTO instructors (course_id, instructor_name) VALUES ('ΤΕΧΝΟΛΟΓΙΑ ΛΟΓΙΣΜΙΚΟΥ   (3205)', 'instructor');
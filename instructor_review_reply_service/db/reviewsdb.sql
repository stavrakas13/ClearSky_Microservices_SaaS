-- for debugging purposes.
DROP TABLE IF EXISTS reviews;

-- reviews made by students added here.
CREATE TABLE IF NOT EXISTS reviews (
  review_id SERIAL PRIMARY KEY,
  student_id VARCHAR(15) NOT NULL,
  course_id VARCHAR(50) NOT NULL,
  exam_period VARCHAR(15) NOT NULL,
  student_message TEXT NOT NULL,
  status VARCHAR(15) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed')),
  instructor_reply_message TEXT DEFAULT NULL,
  instructor_action VARCHAR(25) DEFAULT NULL CHECK (instructor_action IN ('Total Accept', 'Will be considered', 'Denied')),
  review_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  reviewed_at TIMESTAMP DEFAULT NULL
);

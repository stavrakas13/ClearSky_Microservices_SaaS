-- for debugging purposes.
DROP TABLE IF EXISTS reviews;

-- reviews made by students added here.
CREATE TABLE IF NOT EXISTS reviews (
  review_id SERIAL PRIMARY KEY,
  student_id INT NOT NULL,
  course_id INT NOT NULL,
  student_message TEXT NOT NULL,
  status VARCHAR(15) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed')),
  instructor_reply_message TEXT DEFAULT NULL,
  instructor_action VARCHAR(25) DEFAULT NULL CHECK (instructor_action IN ('Total Accept', 'Will be considered', 'Denied')),
  review_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  reviewed_at TIMESTAMP DEFAULT NULL
);

-- test data
INSERT INTO reviews (student_id, course_id, student_message)
VALUES (11111, 22222, 'test');
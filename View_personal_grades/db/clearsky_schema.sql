DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS exams;

/* stores data like 
VALUES ('2024-2025 ΧΕΙΜ 2024', 'ΤΕΧΝΟΛΟΓΙΑ ΛΟΓΙΣΜΙΚΟΥ (3205)', 'NTUA', 'T3205', '{"min": 0, "max": 10}', '[5.0, 40.0, 5.0, 5.0, 5.0, 5.0, 20.0, 5.0, 5.0, 5.0]' );
 */
CREATE TABLE exams (
    exam_date TEXT NOT NULL,
    course_name TEXT NOT NULL,
    course_id TEXT NOT NULL,
    uni_id TEXT NOT NULL,
/*  teacher_id TEXT NOT NULL, -> basically the course_id */
    mark_scale JSONB NOT NULL,
    weights JSONB NOT NULL,
    PRIMARY KEY (exam_date, class_id)
);

/* stores data like
INSERT INTO grades (exam_date, class_id, student_id, question_scores, total_score) VALUES ('2024-2025 ΧΕΙΜ 2024', 'ΤΕΧΝΟΛΟΓΙΑ ΛΟΓΙΣΜΙΚΟΥ (3205)', '03184623', '[8.0, 4.0, 3.0, 8.0, 9.0, 7.0, 10.0, 1.0, 4.0, 8.0]', 6.0);
 */
CREATE TABLE grades (
    exam_date TEXT NOT NULL,
    class_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    question_scores JSONB NOT NULL,
    total_score FLOAT,
    is_finalized BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (exam_date, class_id, student_id),
    FOREIGN KEY (exam_date, class_id)
        REFERENCES exams(exam_date, class_id)
);


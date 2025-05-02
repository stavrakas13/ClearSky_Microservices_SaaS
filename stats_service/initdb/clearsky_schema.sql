DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS exam_students;
DROP TABLE IF EXISTS exams;

CREATE TABLE exams (
    exam_date TEXT NOT NULL,
    class_id TEXT NOT NULL,
    uni_id TEXT NOT NULL,
    teacher_id TEXT NOT NULL,
    mark_scale JSONB NOT NULL,
    weights JSONB NOT NULL,
    PRIMARY KEY (exam_date, class_id)
);

CREATE TABLE grades (
    exam_date TEXT NOT NULL,
    class_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    question_scores JSONB NOT NULL,
    total_score FLOAT,
    PRIMARY KEY (exam_date, class_id, student_id),
    FOREIGN KEY (exam_date, class_id)
        REFERENCES exams(exam_date, class_id)
);

CREATE TABLE grade_distributions (
    class_id TEXT NOT NULL,
    exam_date TEXT NOT NULL,
    category TEXT NOT NULL,
    value INT NOT NULL,
    count INT NOT NULL,
    PRIMARY KEY (class_id, exam_date, category, value)
);

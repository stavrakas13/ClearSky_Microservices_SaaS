-- ==========================================================
--  Πίνακας βαθμολογιών ανά φοιτητή και ερώτημα
-- ==========================================================
CREATE TABLE IF NOT EXISTS grades (
    -- βασική ταυτότητα φοιτητή (PK)
    student_id     TEXT PRIMARY KEY,

    -- μετα-δεδομένα
    name           TEXT        NOT NULL,
    e_mail         TEXT        NOT NULL,
    uni_id         TEXT        NOT NULL DEFAULT 'NTUA',

    -- πληροφορίες εξέτασης
    exam_date      TEXT        NOT NULL,               -- π.χ. '2024-2025 ΧΕΙΜ 2024'
    class_id       TEXT        NOT NULL,               -- κωδικός τμήματος/μαθήματος
    mark_scale     TEXT        NOT NULL DEFAULT '0-10',

    -- συνολικός βαθμός
    total_mark     NUMERIC(10,2) NOT NULL CHECK (total_mark BETWEEN 0 AND 10),

    -- αναλυτικοί βαθμοί ερωτημάτων (NULL αν δεν υπάρχει)
    q01 NUMERIC(10,2) CHECK (q01 BETWEEN 0 AND 10),
    q02 NUMERIC(10,2) CHECK (q02 BETWEEN 0 AND 10),
    q03 NUMERIC(10,2) CHECK (q03 BETWEEN 0 AND 10),
    q04 NUMERIC(10,2) CHECK (q04 BETWEEN 0 AND 10),
    q05 NUMERIC(10,2) CHECK (q05 BETWEEN 0 AND 10),
    q06 NUMERIC(10,2) CHECK (q06 BETWEEN 0 AND 10),
    q07 NUMERIC(10,2) CHECK (q07 BETWEEN 0 AND 10),
    q08 NUMERIC(10,2) CHECK (q08 BETWEEN 0 AND 10),
    q09 NUMERIC(10,2) CHECK (q09 BETWEEN 0 AND 10),
    q10 NUMERIC(10,2) CHECK (q10 BETWEEN 0 AND 10)
);

-- προτεινόμενα indexes για συχνά queries (προαιρετικά):
CREATE INDEX IF NOT EXISTS idx_grades_class ON grades (class_id);
CREATE INDEX IF NOT EXISTS idx_grades_total_mark ON grades (total_mark);


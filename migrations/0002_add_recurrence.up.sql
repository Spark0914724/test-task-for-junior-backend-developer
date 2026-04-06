CREATE TABLE IF NOT EXISTS task_recurrences (
    id          BIGSERIAL PRIMARY KEY,
    type        TEXT    NOT NULL,
    every_n_days INT,
    month_days  INT[]   DEFAULT '{}',
    specific_dates DATE[] DEFAULT '{}',
    parity      TEXT,
    start_date  DATE    NOT NULL,
    end_date    DATE    NOT NULL
);

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS recurrence_id  BIGINT REFERENCES task_recurrences(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS scheduled_date DATE;

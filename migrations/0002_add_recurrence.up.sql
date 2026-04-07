ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS recurrence_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN IF NOT EXISTS recurrence_interval INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS recurrence_day_of_month INTEGER,
ADD COLUMN IF NOT EXISTS recurrence_specific_dates JSONB;

CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_type ON tasks (recurrence_type);
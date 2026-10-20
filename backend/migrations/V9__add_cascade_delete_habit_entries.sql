ALTER TABLE habit_entries
DROP CONSTRAINT habit_entries_habit_id_fkey;

ALTER TABLE habit_entries
ADD CONSTRAINT habit_entries_habit_id_fkey
FOREIGN KEY (habit_id) REFERENCES habits(id) ON DELETE CASCADE;
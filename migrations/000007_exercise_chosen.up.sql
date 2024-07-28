
-- exercise_chosen_table.sql
-- Create exercise_chosen table
CREATE TABLE exercise_chosen (
                                 id SERIAL PRIMARY KEY,
                                 exercise_id INT,
                                 user_id INT,
                                 done BOOLEAN,
                                 created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
                                 type varchar(40),
                                 CONSTRAINT fk_exercise_chosen_exercise FOREIGN KEY (exercise_id) REFERENCES exercise(id),
                                 CONSTRAINT fk_exercise_chosen_user FOREIGN KEY (user_id) REFERENCES users(id)
);

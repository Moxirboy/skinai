-- program_chosen_table.sql
-- Create program_chosen table
CREATE TABLE program_chosen (
                                id SERIAL PRIMARY KEY,
                                program_id INT NOT NULL,
                                user_id INT NOT NULL,
                                CONSTRAINT fk_program_chosen_program FOREIGN KEY (program_id) REFERENCES programs(id),
                                CONSTRAINT fk_program_chosen_user FOREIGN KEY (user_id) REFERENCES users(id)
);

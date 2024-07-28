-- user_info_table.sql
-- Create user_info table
CREATE TABLE user_info (
                           id SERIAL PRIMARY KEY,
                           user_id INT NOT NULL,
                           name VARCHAR(255),
                            
                           created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
                           updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
                           deleted_at TIMESTAMP,
                           CONSTRAINT fk_user_info_user FOREIGN KEY (user_id) REFERENCES users(id)
);

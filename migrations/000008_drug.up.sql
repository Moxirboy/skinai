-- drug_table.sql
-- Create drug table
CREATE TABLE drug (
                      id SERIAL PRIMARY KEY,
                      name VARCHAR(255),
                      description VARCHAR(255),
                      manufacturer VARCHAR(255),
                      reciept VARCHAR(255)
);

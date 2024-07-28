-- programs_table.sql
-- Create programs table
CREATE TYPE program_type AS ENUM ('weight_loss', 'stress_work');
CREATE TYPE pro_type AS ENUM ('recommended', 'personal');

CREATE TABLE programs (
                          id SERIAL PRIMARY KEY,
                          ageUp int  NOT NULL,
                          ageDown int NOT NULL,
                          bmiUp decimal  NOT NULL,
                          bmiDown decimal  NOT NULL,
                          type program_type NOT NULL,
                          pro_type pro_type NOT NULL
);

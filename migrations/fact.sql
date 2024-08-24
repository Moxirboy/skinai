create table fact (
    id int autoincrement,
    title text,
    content text,
    image text,
    number_question int
);

create table question (
    id int autoincrement,
    fact_id int,
    question text
);

create table choices (
    id int autoincrement,
    question_id int,
    content text,
    is_true boolean
);


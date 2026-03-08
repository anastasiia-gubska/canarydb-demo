DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
    full_name TEXT 
);

INSERT INTO users (full_name) VALUES 
('Isaac Newton'),
('William Shakespeare'),
('Charles Darwin');
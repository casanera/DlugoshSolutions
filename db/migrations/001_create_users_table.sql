CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,         -- SERIAL - автоинкрементный целочисленный ключ
    name VARCHAR(100) NOT NULL,    -- VARCHAR(100) - строка до 100 символов, NOT NULL - обязательное поле
    email VARCHAR(100) UNIQUE NOT NULL -- UNIQUE - значение должно быть уникальным в таблице
);
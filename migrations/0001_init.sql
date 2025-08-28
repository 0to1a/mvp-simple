
-- +goose Up
CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL CONSTRAINT users_email_unique UNIQUE,
    otp        VARCHAR(8) DEFAULT '000000'::CHARACTER VARYING NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE companies (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    address    VARCHAR(500),
    phone      VARCHAR(50),
    email      VARCHAR(255),
    tax_id     VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE user_companies (
    user_id    INTEGER NOT NULL CONSTRAINT user_companies_user_id_users_id_fk 
               REFERENCES users ON DELETE CASCADE,
    company_id INTEGER NOT NULL CONSTRAINT user_companies_company_id_companies_id_fk 
               REFERENCES companies ON DELETE CASCADE,
    is_admin   BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT user_companies_user_id_company_id_pk PRIMARY KEY (user_id, company_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_companies;
DROP TABLE IF EXISTS companies;
DROP TABLE IF EXISTS users;

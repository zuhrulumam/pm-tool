-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS auth;

-- Create users table
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    name VARCHAR(100),
    google_id VARCHAR(255) UNIQUE,
    avatar_url VARCHAR(512),
    email_verified BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);


CREATE UNIQUE INDEX idx_users_email ON auth.users (email);
CREATE UNIQUE INDEX idx_users_google_id ON auth.users (google_id);
CREATE INDEX idx_users_is_active ON auth.users (is_active);

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION auth.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_updated_at_on_users ON auth.users;
CREATE TRIGGER set_updated_at_on_users
BEFORE UPDATE ON auth.users
FOR EACH ROW
EXECUTE FUNCTION auth.update_updated_at_column();

-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS project;

-- Create projects table
CREATE TABLE IF NOT EXISTS project.projects (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#3B82F6',
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- Foreign key: user_id -> users(id)
ALTER TABLE project.projects ADD CONSTRAINT fk_projects_user_id_users_id
    FOREIGN KEY (user_id) REFERENCES auth.users(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE;


CREATE INDEX idx_projects_user_id ON project.projects (user_id);
CREATE INDEX idx_projects_user_id_is_archived ON project.projects (user_id, is_archived);
CREATE INDEX idx_projects_created_at ON project.projects (created_at);

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION project.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_updated_at_on_projects ON project.projects;
CREATE TRIGGER set_updated_at_on_projects
BEFORE UPDATE ON project.projects
FOR EACH ROW
EXECUTE FUNCTION project.update_updated_at_column();

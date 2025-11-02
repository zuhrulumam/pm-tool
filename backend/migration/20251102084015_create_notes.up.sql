-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS project;

-- Create enum type if not exists

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notes_content_type_enum') THEN
        CREATE TYPE notes_content_type_enum AS ENUM ('text', 'audio_transcription');
    END IF;
END$$;

-- Create notes table
CREATE TABLE IF NOT EXISTS project.notes (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    content_type notes_content_type_enum NOT NULL DEFAULT 'text',
    audio_url VARCHAR(512),
    is_pinned BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- Foreign key: project_id -> projects(id)
ALTER TABLE project.notes ADD CONSTRAINT fk_notes_project_id_projects_id
    FOREIGN KEY (project_id) REFERENCES project.projects(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE;


CREATE INDEX idx_notes_project_id ON project.notes (project_id);
CREATE INDEX idx_notes_project_id_created_at ON project.notes (project_id, created_at);
CREATE INDEX idx_notes_project_id_is_pinned ON project.notes (project_id, is_pinned);
CREATE INDEX idx_notes_created_at ON project.notes (created_at);

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION project.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_updated_at_on_notes ON project.notes;
CREATE TRIGGER set_updated_at_on_notes
BEFORE UPDATE ON project.notes
FOR EACH ROW
EXECUTE FUNCTION project.update_updated_at_column();

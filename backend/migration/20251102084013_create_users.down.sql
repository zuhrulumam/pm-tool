-- +migrate Down
-- Drop table if exists
DROP TABLE IF EXISTS auth.users CASCADE;

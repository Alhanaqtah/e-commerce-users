DROP INDEX IF EXISTS idx_local_credentials_email;

DROP TABLE local_credentials IF EXISTS;
DROP TABLE users IF EXISTS;

DROP ENUM IF EXISTS role;

DROP EXTENSION "uuid-ossp" IF EXISTS;
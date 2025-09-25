-- Create separate roles for migrator and app with least privilege
DO $$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'flight_migrator') THEN
      CREATE ROLE flight_migrator LOGIN PASSWORD 'migrator';
   END IF;
   IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'flight_app') THEN
      CREATE ROLE flight_app LOGIN PASSWORD 'app';
   END IF;
END$$;

-- Grant privileges to migrator
GRANT ALL PRIVILEGES ON DATABASE flight TO flight_migrator;
GRANT USAGE, CREATE ON SCHEMA public TO flight_migrator;

-- Grant connect/use to app; table privileges will be granted post-migration
GRANT CONNECT ON DATABASE flight TO flight_app;
GRANT USAGE ON SCHEMA public TO flight_app;

-- Ensure future objects created by migrator grant app DML automatically
ALTER DEFAULT PRIVILEGES FOR ROLE flight_migrator IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO flight_app;
ALTER DEFAULT PRIVILEGES FOR ROLE flight_migrator IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO flight_app;

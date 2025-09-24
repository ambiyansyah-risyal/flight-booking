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

-- Grant connect/use to app; table privileges will be granted post-migration
GRANT CONNECT ON DATABASE flight TO flight_app;


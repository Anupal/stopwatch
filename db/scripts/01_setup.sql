-- 1. Enable extensions
CREATE EXTENSION IF NOT EXISTS postgis;
-- handles names like Café
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. Speeds up all index builds
SET maintenance_work_mem = '1GB';
SET max_parallel_maintenance_workers = 4;
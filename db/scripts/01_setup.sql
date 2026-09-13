-- 1. Enable postgis extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- 2. Speeds up all index builds
SET maintenance_work_mem = '1GB';
SET max_parallel_maintenance_workers = 4;
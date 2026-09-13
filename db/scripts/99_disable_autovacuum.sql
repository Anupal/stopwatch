VACUUM FREEZE ANALYZE;
ALTER SYSTEM SET autovacuum = off;
SELECT pg_reload_conf()
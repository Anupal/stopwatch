BEGIN;
CREATE TABLE agency (
    agency_id TEXT PRIMARY KEY,
    agency_name TEXT NOT NULL,
    agency_url TEXT NOT NULL,
    agency_timezone TEXT NOT NULL
) WITH (fillfactor = 100);

COPY agency(
    agency_id,
    agency_name,
    agency_url,
    agency_timezone
) FROM '/gtfs_data/agency.txt'
WITH (FORMAT csv, HEADER true, FREEZE);

COMMIT;

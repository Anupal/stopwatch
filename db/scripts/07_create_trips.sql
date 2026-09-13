BEGIN;

CREATE TABLE trips (
    route_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    trip_id TEXT,
    trip_headsign TEXT,
    trip_short_name TEXT,
    direction_id INT,
    block_id TEXT,
    shape_id TEXT
) WITH (fillfactor = 100);

COPY trips(
    route_id, service_id, trip_id, trip_headsign,
    trip_short_name, direction_id, block_id, shape_id
) FROM '/gtfs_data/trips.txt'
WITH (FORMAT csv, HEADER true, FREEZE);

ALTER TABLE trips ADD PRIMARY KEY (trip_id);

COMMIT;
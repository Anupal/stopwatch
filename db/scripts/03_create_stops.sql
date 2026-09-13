CREATE TABLE stops_loaded (
    stop_id TEXT NOT NULL,
    stop_code TEXT NOT NULL,
    stop_name TEXT NOT NULL,
    stop_desc TEXT,
    stop_lat DOUBLE PRECISION NOT NULL,
    stop_lon DOUBLE PRECISION NOT NULL,
    zone_id TEXT,
    stop_url TEXT,
    location_type INT,
    parent_station TEXT
) WITH (fillfactor = 100);

COPY stops_loaded(
    stop_id, stop_code, stop_name, stop_desc,
    stop_lat, stop_lon, zone_id, stop_url,
    location_type, parent_station
) FROM '/gtfs_data/stops.txt'
WITH (FORMAT csv, HEADER true);

CREATE TABLE stops WITH (fillfactor = 100) AS
SELECT *, ST_SetSRID(ST_MakePoint(stop_lon, stop_lat), 4326)::geometry(Point,4326) AS geom
FROM stops_loaded;

DROP TABLE stops_loaded;

ALTER TABLE stops ADD PRIMARY KEY (stop_id);

VACUUM FREEZE stops;
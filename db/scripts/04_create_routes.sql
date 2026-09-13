BEGIN;
CREATE TABLE routes (
    route_id TEXT NOT NULL,
    agency_id TEXT NOT NULL,
    route_short_name TEXT NOT NULL,
    route_long_name TEXT NOT NULL,
    route_desc TEXT,
    route_type INT NOT NULL,
    route_url TEXT,
    route_color TEXT,
    route_text_color TEXT
) WITH (fillfactor = 100);

COPY routes(
    route_id, agency_id, route_short_name,
    route_long_name, route_desc, route_type,
    route_url, route_color, route_text_color
) FROM '/gtfs_data/routes.txt'
WITH (FORMAT csv, HEADER true, FREEZE);

ALTER TABLE routes ADD PRIMARY KEY (route_id);

COMMIT;
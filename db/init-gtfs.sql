-- 1. Enable postgis extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- 2. Create Core Tables
CREATE TABLE agency (
    agency_id TEXT PRIMARY KEY,
    agency_name TEXT NOT NULL,
    agency_url TEXT NOT NULL,
    agency_timezone TEXT NOT NULL
);

CREATE TABLE stops (
    stop_id TEXT PRIMARY KEY,
    stop_code TEXT,
    stop_name TEXT NOT NULL,
    stop_desc TEXT,
    stop_lat DOUBLE PRECISION NOT NULL,
    stop_lon DOUBLE PRECISION NOT NULL,
    zone_id TEXT,
    stop_url TEXT,
    location_type INT,
    parent_station TEXT,
    geom geometry(Point, 4326)
);

CREATE TABLE routes (
    route_id TEXT PRIMARY KEY,
    agency_id TEXT REFERENCES agency(agency_id),
    route_short_name TEXT,
    route_long_name TEXT,
    route_desc TEXT,
    route_type INT NOT NULL,
    route_url TEXT,
    route_color TEXT,
    route_text_color TEXT
);

CREATE TABLE calendar (
    service_id TEXT PRIMARY KEY,
    monday INT NOT NULL,
    tuesday INT NOT NULL,
    wednesday INT NOT NULL,
    thursday INT NOT NULL,
    friday INT NOT NULL,
    saturday INT NOT NULL,
    sunday INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
);

CREATE TABLE calendar_dates (
    service_id TEXT NOT NULL,
    date DATE NOT NULL,
    exception_type INT NOT NULL, -- 1 = added, 2 = removed
    PRIMARY KEY (service_id, date)
);

CREATE TABLE trips (
    route_id TEXT REFERENCES routes(route_id),
    service_id TEXT NOT NULL,
    trip_id TEXT PRIMARY KEY,
    trip_headsign TEXT,
    trip_short_name TEXT,
    direction_id INT,
    block_id TEXT,
    shape_id TEXT
);

-- Note: Times are Stored as TEXT due to GTFS times exceeding 24:00:00
-- (e.g. 25:30:00 for late night service)
CREATE TABLE stop_times (
    trip_id TEXT REFERENCES trips(trip_id),
    arrival_time TEXT NOT NULL,
    departure_time TEXT NOT NULL,
    stop_id TEXT REFERENCES stops(stop_id),
    stop_sequence INT NOT NULL,
    stop_headsign TEXT,
    pickup_type INT,
    drop_off_type INT,
    timepoint INT,
    PRIMARY KEY (trip_id, stop_sequence)
);


-- 3. Import CSV/TXT Files into Tables

\copy agency(agency_id, agency_name, agency_url, agency_timezone) FROM '/gtfs_data/agency.txt' WITH (FORMAT csv, HEADER true);
\copy stops(stop_id, stop_code, stop_name, stop_desc, stop_lat, stop_lon, zone_id, stop_url, location_type, parent_station) FROM '/gtfs_data/stops.txt' WITH (FORMAT csv, HEADER true);
\copy routes(route_id, agency_id, route_short_name, route_long_name, route_desc, route_type, route_url, route_color, route_text_color) FROM '/gtfs_data/routes.txt' WITH (FORMAT csv, HEADER true);
\copy calendar(service_id, monday, tuesday, wednesday, thursday, friday, saturday, sunday, start_date, end_date) FROM '/gtfs_data/calendar.txt' WITH (FORMAT csv, HEADER true);
\copy calendar_dates(service_id, date, exception_type) FROM '/gtfs_data/calendar_dates.txt' WITH (FORMAT csv, HEADER true);
\copy trips(route_id, service_id, trip_id, trip_headsign, trip_short_name, direction_id, block_id, shape_id) FROM '/gtfs_data/trips.txt' WITH (FORMAT csv, HEADER true);
\copy stop_times(trip_id, arrival_time, departure_time, stop_id, stop_sequence, stop_headsign, pickup_type, drop_off_type, timepoint) FROM '/gtfs_data/stop_times.txt' WITH (FORMAT csv, HEADER true);

-- 4. Populate PostGIS geometry column for stops

UPDATE stops 
SET geom = ST_SetSRID(ST_MakePoint(stop_lon, stop_lat), 4326)
WHERE geom IS NULL;

CREATE INDEX idx_stops_geom ON stops USING gist (geom);

-- 5. Create a view and function for upcoming arrivals for a stop
-- View for today's active schedule for a stop
CREATE OR REPLACE VIEW v_today_arrivals AS
WITH active_services AS (
    SELECT c.service_id 
    FROM calendar c
    WHERE CURRENT_DATE BETWEEN c.start_date AND c.end_date
      AND (
          (EXTRACT(DOW FROM CURRENT_DATE) = 0 AND c.sunday = 1) OR
          (EXTRACT(DOW FROM CURRENT_DATE) = 1 AND c.monday = 1) OR
          (EXTRACT(DOW FROM CURRENT_DATE) = 2 AND c.tuesday = 1) OR
          (EXTRACT(DOW FROM CURRENT_DATE) = 3 AND c.wednesday = 1) OR
          (EXTRACT(DOW FROM CURRENT_DATE) = 4 AND c.thursday = 1) OR
          (EXTRACT(DOW FROM CURRENT_DATE) = 5 AND c.friday = 1) OR
          (EXTRACT(DOW FROM CURRENT_DATE) = 6 AND c.saturday = 1)
      )
      AND c.service_id NOT IN (
          SELECT cd.service_id 
          FROM calendar_dates cd 
          WHERE cd.date = CURRENT_DATE AND cd.exception_type = 2
      )
    UNION
    SELECT cd.service_id 
    FROM calendar_dates cd
    WHERE cd.date = CURRENT_DATE AND cd.exception_type = 1
)
SELECT 
    t.trip_id,
    st.stop_id,
    s.stop_name,
    st.arrival_time,
    st.departure_time,
    st.arrival_time::interval AS arrival_interval,
    r.route_short_name,
    r.route_long_name,
    t.trip_headsign
FROM stop_times st
JOIN trips t ON st.trip_id = t.trip_id
JOIN routes r ON t.route_id = r.route_id
JOIN stops s ON st.stop_id = s.stop_id
WHERE t.service_id IN (SELECT service_id FROM active_services);

-- Function to get windowed Arrivals (-p_minutes_behind to +p_minutes_ahead)
CREATE OR REPLACE FUNCTION get_upcoming_arrivals(
    p_stop_id TEXT,
    p_minutes_ahead INT DEFAULT 30,
    p_minutes_behind INT DEFAULT 15
)
RETURNS TABLE (
    trip_id TEXT,
    arrival_time TEXT,
    departure_time TEXT,
    route_short_name TEXT,
    route_long_name TEXT,
    trip_headsign TEXT,
    stop_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        v.trip_id,
        v.arrival_time,
        v.departure_time,
        v.route_short_name,
        v.route_long_name,
        v.trip_headsign,
        v.stop_name
    FROM v_today_arrivals v
    WHERE v.stop_id = p_stop_id
      AND v.arrival_interval >= LOCALTIME::interval - (p_minutes_behind || ' minutes')::interval
      AND v.arrival_interval <= LOCALTIME::interval + (p_minutes_ahead || ' minutes')::interval
    ORDER BY v.arrival_interval ASC;
END;
$$ LANGUAGE plpgsql STABLE;

-- 6. Create function for nearest stop by agency
CREATE OR REPLACE FUNCTION get_nearest_stops_by_agency(
    p_agency_id TEXT,
    p_lat DOUBLE PRECISION,
    p_lon DOUBLE PRECISION,
    p_limit INT DEFAULT 10,
    p_max_distance_meters DOUBLE PRECISION DEFAULT 5000
)
RETURNS TABLE (
    stop_id TEXT,
    stop_name TEXT,
    stop_lat DOUBLE PRECISION,
    stop_lon DOUBLE PRECISION,
    distance_meters DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    WITH agency_stops AS (
        -- Get unique stops served by routes belonging to the specified agency
        SELECT DISTINCT 
            s.stop_id, 
            s.stop_name, 
            s.stop_lat, 
            s.stop_lon, 
            s.geom
        FROM stops s
        JOIN stop_times st ON s.stop_id = st.stop_id
        JOIN trips t ON st.trip_id = t.trip_id
        JOIN routes r ON t.route_id = r.route_id
        WHERE r.agency_id = p_agency_id
    )
    SELECT 
        ast.stop_id,
        ast.stop_name,
        ast.stop_lat,
        ast.stop_lon,
        ROUND(
            ST_Distance(
                ast.geom::geography, 
                ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography
            )::numeric, 2
        )::DOUBLE PRECISION AS distance_meters
    FROM agency_stops ast
    WHERE ST_DWithin(
        ast.geom::geography, 
        ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography, 
        p_max_distance_meters
    )
    ORDER BY ast.geom::geography <-> ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql STABLE;
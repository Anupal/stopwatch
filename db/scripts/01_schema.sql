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

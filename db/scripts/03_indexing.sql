-- 1. Populate PostGIS geometry column for stops

UPDATE stops 
SET geom = ST_SetSRID(ST_MakePoint(stop_lon, stop_lat), 4326)
WHERE geom IS NULL;

CREATE INDEX idx_stops_geom ON stops USING gist (geom);

-- 2. Enables sub-millisecond lookups for get_upcoming_arrivals()
-- Primary Key is (trip_id, stop_sequence), which cannot index queries filtering by stop_id alone.
CREATE INDEX idx_stop_times_stop_id ON stop_times(stop_id);

-- 3. Enables fast joins for active service resolution in v_today_arrivals
-- Primary Key is (trip_id), so filtering trips by active service_id requires a full scan without this.
CREATE INDEX idx_trips_service_id ON trips(service_id);

-- 4. Speeds up joining trips to routes (agency filtering)
CREATE INDEX idx_trips_route_id ON trips(route_id);

-- 5. Speeds up get_nearest_stops_by_agency() when searching by agency_id
CREATE INDEX idx_routes_agency_id ON routes(agency_id);

-- 6. Speeds up checking date exceptions in active service CTEs
CREATE INDEX idx_calendar_dates_date ON calendar_dates(date);
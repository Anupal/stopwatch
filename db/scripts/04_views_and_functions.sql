-- 1. Create a view and function for upcoming arrivals for a stop
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
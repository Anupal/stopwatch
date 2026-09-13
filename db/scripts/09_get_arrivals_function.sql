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
    st.arrival_interval,
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
    trip_headsign TEXT
) AS $$
DECLARE
    v_now       INTERVAL := LOCALTIME::interval;
    v_lo        INTERVAL := LOCALTIME::interval - (p_minutes_behind || ' minutes')::interval;
    v_hi        INTERVAL := LOCALTIME::interval + (p_minutes_ahead  || ' minutes')::interval;
BEGIN
    RETURN QUERY
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
    ),
    windowed_stop_times AS (
        -- This is the piece that hits idx_stop_times_stop_arrival:
        -- an index range scan on (stop_id, arrival_interval), nothing else.
        SELECT st.trip_id, st.arrival_time, st.departure_time
        FROM stop_times st
        WHERE st.stop_id = p_stop_id
          AND st.arrival_interval >= v_lo
          AND st.arrival_interval <= v_hi
    )
    SELECT 
        wst.trip_id,
        wst.arrival_time,
        wst.departure_time,
        r.route_short_name,
        r.route_long_name,
        t.trip_headsign
    FROM windowed_stop_times wst
    JOIN trips  t ON wst.trip_id = t.trip_id
    JOIN routes r ON t.route_id = r.route_id
    WHERE t.service_id IN (SELECT service_id FROM active_services)
    ORDER BY wst.arrival_time::interval ASC;
END;
$$ LANGUAGE plpgsql STABLE;
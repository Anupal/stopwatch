CREATE MATERIALIZED VIEW mv_agency_stops AS
SELECT DISTINCT r.agency_id, s.stop_id, s.stop_name, s.geom
FROM stops s
JOIN stop_times st ON s.stop_id = st.stop_id
JOIN trips t ON st.trip_id = t.trip_id
JOIN routes r ON t.route_id = r.route_id;

CREATE INDEX idx_mv_agency_stops_agency ON mv_agency_stops (agency_id);
CREATE INDEX idx_mv_agency_stops_geom ON mv_agency_stops USING gist (geom);

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
    SELECT
        mas.stop_id,
        mas.stop_name,
        ST_Y(mas.geom)::DOUBLE PRECISION AS stop_lat,
        ST_X(mas.geom)::DOUBLE PRECISION AS stop_lon,
        ROUND(
            ST_Distance(
                mas.geom::geography,
                ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography
            )::numeric, 2
        )::DOUBLE PRECISION AS distance_meters
    FROM mv_agency_stops mas
    WHERE mas.agency_id = p_agency_id
      AND ST_DWithin(
            mas.geom::geography,
            ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography,
            p_max_distance_meters
          )
    ORDER BY mas.geom::geography <-> ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql STABLE;
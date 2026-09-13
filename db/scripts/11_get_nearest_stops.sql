CREATE OR REPLACE FUNCTION get_nearest_stops_by_agency(
    p_agency_ids TEXT[],
    p_lat DOUBLE PRECISION,
    p_lon DOUBLE PRECISION,
    p_max_distance_meters DOUBLE PRECISION DEFAULT 500
)
RETURNS TABLE (
    stop_id TEXT,
    stop_name TEXT,
    stop_code TEXT,
    stop_lat DOUBLE PRECISION,
    stop_lon DOUBLE PRECISION,
    agency_id TEXT,
    agency_name TEXT,
    agency_url TEXT,
    distance_meters DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        mas.stop_id,
        mas.stop_name,
        mas.stop_code,
        ST_Y(mas.geom)::DOUBLE PRECISION AS stop_lat,
        ST_X(mas.geom)::DOUBLE PRECISION AS stop_lon,
        mas.agency_id,
        mas.agency_name,
        mas.agency_url,
        ROUND(
            ST_Distance(
                mas.geom::geography,
                ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography
            )::numeric, 2
        )::DOUBLE PRECISION AS distance_meters
    FROM mv_agency_stops mas
    WHERE mas.agency_id = ANY(p_agency_ids)
      AND ST_DWithin(
            mas.geom::geography,
            ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography,
            p_max_distance_meters
          )
    ORDER BY mas.geom::geography <-> ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography;
END;
$$ LANGUAGE plpgsql STABLE;
CREATE MATERIALIZED VIEW mv_agency_stops AS
SELECT DISTINCT
    r.agency_id,
    a.agency_name,
    a.agency_url,
    s.stop_id,
    s.stop_code,
    s.stop_name,
    lower(unaccent(s.stop_name)) AS normalized_name,
    s.stop_lat,
    s.stop_lon,
    s.geom
FROM stops s
JOIN stop_times st ON s.stop_id = st.stop_id
JOIN trips t ON st.trip_id = t.trip_id
JOIN routes r ON t.route_id = r.route_id
JOIN agency a ON r.agency_id = a.agency_id;

CREATE INDEX idx_mv_agency_stops_agency ON mv_agency_stops (agency_id);
CREATE INDEX stops_name_trgm_idx ON mv_agency_stops USING gin (normalized_name gin_trgm_ops);
CREATE INDEX idx_mv_agency_stops_geom ON mv_agency_stops USING gist (geom);

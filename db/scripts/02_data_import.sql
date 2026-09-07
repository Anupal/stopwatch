-- 1. Import CSV/TXT Files into Tables

\copy agency(agency_id, agency_name, agency_url, agency_timezone) FROM '/gtfs_data/agency.txt' WITH (FORMAT csv, HEADER true);
\copy stops(stop_id, stop_code, stop_name, stop_desc, stop_lat, stop_lon, zone_id, stop_url, location_type, parent_station) FROM '/gtfs_data/stops.txt' WITH (FORMAT csv, HEADER true);
\copy routes(route_id, agency_id, route_short_name, route_long_name, route_desc, route_type, route_url, route_color, route_text_color) FROM '/gtfs_data/routes.txt' WITH (FORMAT csv, HEADER true);
\copy calendar(service_id, monday, tuesday, wednesday, thursday, friday, saturday, sunday, start_date, end_date) FROM '/gtfs_data/calendar.txt' WITH (FORMAT csv, HEADER true);
\copy calendar_dates(service_id, date, exception_type) FROM '/gtfs_data/calendar_dates.txt' WITH (FORMAT csv, HEADER true);
\copy trips(route_id, service_id, trip_id, trip_headsign, trip_short_name, direction_id, block_id, shape_id) FROM '/gtfs_data/trips.txt' WITH (FORMAT csv, HEADER true);
\copy stop_times(trip_id, arrival_time, departure_time, stop_id, stop_sequence, stop_headsign, pickup_type, drop_off_type, timepoint) FROM '/gtfs_data/stop_times.txt' WITH (FORMAT csv, HEADER true);

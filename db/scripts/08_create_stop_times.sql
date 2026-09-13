BEGIN;

CREATE TABLE stop_times (
    trip_id TEXT NOT NULL,
    arrival_time TEXT NOT NULL,
    departure_time TEXT NOT NULL,
    stop_id TEXT NOT NULL,
    stop_sequence INT NOT NULL,
    stop_headsign TEXT,
    pickup_type INT,
    drop_off_type INT,
    timepoint INT
) WITH (fillfactor = 100);

COPY stop_times(
    trip_id, arrival_time, departure_time, stop_id,
    stop_sequence, stop_headsign, pickup_type,
    drop_off_type, timepoint
) FROM '/gtfs_data/stop_times.txt'
WITH (FORMAT csv, HEADER true, FREEZE);

ALTER TABLE stop_times ADD PRIMARY KEY (trip_id, stop_sequence);

-- TODO: follow claude
ALTER TABLE stop_times ADD COLUMN arrival_interval INTERVAL;
UPDATE stop_times SET arrival_interval = arrival_time::interval;
CREATE INDEX idx_stop_times_stop_arrival ON stop_times (stop_id, arrival_interval) WITH (fillfactor = 100);

COMMIT;
BEGIN;

CREATE TABLE calendar_dates (
    service_id TEXT NOT NULL,
    date DATE NOT NULL,
    exception_type INT NOT NULL -- 1 = added, 2 = removed
) WITH (fillfactor = 100);

COPY calendar_dates(
    service_id,
    date,
    exception_type
) FROM '/gtfs_data/calendar_dates.txt'
WITH (FORMAT csv, HEADER true, FREEZE);

ALTER TABLE calendar_dates ADD PRIMARY KEY (service_id, date);

COMMIT;
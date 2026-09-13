BEGIN;

CREATE TABLE calendar (
    service_id TEXT NOT NULL,
    monday INT NOT NULL,
    tuesday INT NOT NULL,
    wednesday INT NOT NULL,
    thursday INT NOT NULL,
    friday INT NOT NULL,
    saturday INT NOT NULL,
    sunday INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
) WITH (fillfactor = 100);

COPY calendar(
    service_id, monday, tuesday, wednesday,
    thursday, friday, saturday, sunday,
    start_date, end_date
) FROM '/gtfs_data/calendar.txt'
WITH (FORMAT csv, HEADER true, FREEZE);

ALTER TABLE calendar ADD PRIMARY KEY (service_id);

COMMIT;
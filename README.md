# stopwatch
 

### GTFS DB
Download the GTFS data ZIP from [Ireland NTA website](https://developer.nationaltransport.ie/) and place the extracted files in `db/gtfs_data`.
```sh
# Start DB
docker compose up -d

# Execute psql commands
docker exec -it gtfs-postgres psql -U gtfs -d gtfs_db
```
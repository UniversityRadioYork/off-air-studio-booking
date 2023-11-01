# Off-Air Studio Booking

This is designed as a replacement to the booking spreadsheet, because this is actually a calendar.

## Running

The app runs on a Postgres database, the tables can be created with `schema.sql`. Set the environment variables `DBHOST`, `DBPORT`, `DBUSER`, `DBPASS`, `DBNAME` to run the Go server. You'll also need a `.myradio.key` file. Alternatively, use `deploy.sh` as inspiration for running in Docker. In URY, Jenkins uses `deploy.sh` to deploy when pushed to `main`.


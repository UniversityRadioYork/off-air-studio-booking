# Off-Air Studio Booking

This is designed as a replacement to the booking spreadsheet, because this is actually a calendar.

## Running

The app runs on a Postgres database, the tables can be created with `schema.sql`. You'll need to set the below environment variables. You'll also need a `.myradio.key` file. Alternatively, use `deploy.sh` as inspiration for running in Docker. In URY, Jenkins uses `deploy.sh` to deploy when pushed to `main`.

### Environment Variables

- `DBHOST`
- `DBNAME`
- `DBPASS`
- `DBPORT`
- `DBUSER`
- `HOST` (i.e. `http://localhost:8080`)
- `MYRADIO_API_SERVER`
- `MYRADIO_SIGNING_KEY`

> [!TIP]
> Users with a Computing team officership can use the `/cacheview` and `/flush` endpoints to view and flush cached objects respectively.

> [!TIP]
> Training sessions added automatically because they were created in MyRadio will have a 📻 symbol attached.
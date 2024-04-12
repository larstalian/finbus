# Finbus

Finbus is a server application designed to provide real-time tracking of Finnish buses using MQTT protocol, with data
accessibility extended through a REST API and a WebSocket interface.

## Features

- **Real-Time Bus Tracking:** Leveraging MQTT for live updates on bus locations.
- **Accessible Data Endpoints:** REST API endpoints for historical data retrieval and live data via WebSocket.

## Endpoints

### GET /api/get-busses

This endpoint returns all the busses that are close based on your calculated geohash from the posted latitude and
longitude.

### POST /api/stops/get-busses

Gets all buses that has the stops as NextStop. Currently only returning the Vehicle ID`s.

### Websocket ws/bus-updates

This endpoint is a websocket that sends updates on the busses that are close to the calculated geohash from the posted
latitude and longitude.

### How to install and run

1. Clone the repository
2. run

```bash
docker-compose up
```

# Test

I have written e2e tests to test the functionality. These tests are not using mocks to demonstrate the functionality of
the app. To run the tests, run the following command in the root folder:

Therefore, you must ***ONLY*** start the docker service for the database to be available!

Again, the tests are to demonstrate the functionality of the app and not to be used in a production environment as we
are actually creating data in the database.

```bash
godotenv -f ./.env go test ./tests/
```

This sets up the environment variables and runs the tests.

If you don`t have godotenv installed, run the following command:

```bash
go install github.com/joho/godotenv/cmd/godotenv@latest
```

## Further Work

Unit tests.

use proto

better error handling for database not running

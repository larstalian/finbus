# Finus
Server for listening to the finnish busses through MQTT and stores the data in a database.

The app has an API endpoint for publishing your position, which starts listening to the corresponding geohash head in the MQTT server. This stores the data in a influxDB database.

 

## How to install and run
1. Clone the repository
2. run 
```bash
docker-compose up
```


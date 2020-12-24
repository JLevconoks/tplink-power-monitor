An app which scrapes power metrics of TP-Link HS110 Smart Socket and stores it in InfluxDB for home monitoring dashboard. 

Original idea from https://github.com/softScheck/tplink-smartplug

Required ENV vars:
- TARGETS: Comma delimited named target urls with names, e.g "Socket1:192.168.0.100:9999"
- INFLUX_URL: InfluxDB url with port
- BUCKET_NAME: InfluxDB bucket name
- ORG_NAME: InfluxDB org name
- TOKEN: InfluxDB bucket access token


Running in Docker: 
```shell
docker build -t jlevconoks/tplink-power-monitor .

docker run --name tplink-power-monitor \
-e TARGETS=Socket1:192.168.0.100:9999 \
-e INFLUX_URL=127.0.0.1:9999 \
-e BUCKET_NAME=powermonitor \
-e ORG_NAME=home \
-e TOKEN=sometoken \
jlevconoks/tplink-power-monitor
```
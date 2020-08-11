# goqtt-influx

Example connecting a MQTT client to an InfluxDB database using `goqtt`.

## configuration

Configuration is done through environment variables. The InfluxDB server and bucket, as well as the MQTT broker and topic, should be set using the following environment variable names.

```text
export INFLUX_HOST=http://192.168.1.210:8086
export INFLUX_BUCKET=goqtt-bucket
export MQTT_HOST=192.168.1.243
export MQTT_TOPIC=arduino/goqtt
```
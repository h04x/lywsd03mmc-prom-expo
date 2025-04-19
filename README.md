# LYWSD03MMC prometheus exporter

Currently only supported [pvvx/ATC_MiThermometer](https://github.com/pvvx/ATC_MiThermometer) firmware. 
Custom advertisement format

## Requirements 

`Bluez` used on linux box, it must be installed

## Building

`make build`

## Usage
 See all args `lywsd03mmc-prom-expo -h`  
Minimal run, pass thermometers
```
./lywsd03mmc-prom-expo --dev AA:BB:CC:DD:EE:FF --dev 11:22:33:44:55:66
```

## Installation systemd
Build and copy config files 
```
sudo make install
```  
Edit config file 
```
nano /etc/default/lywsd03mmc-prom-expo
```

Enable and start service  
```
systemctl enable lywsd03mmc-prom-expo
systemctl start lywsd03mmc-prom-expo
```
Check status  
```
systemctl status lywsd03mmc-prom-expo
```

## Installation via docker
Buld image
```
docker buildx build -t lywsd03mmc-prom-expo .
```
Debug run. Do not forget replace `--dev` to real devices
```
docker run -it --rm --network=host --privileged \
-v /var/run/dbus/:/var/run/dbus/:z lywsd03mmc-prom-expo \
--listen-address 127.0.0.1:8091 --bt-scan-timeout 31s \
--dev 00:11:22:33:44:55 --dev AA:BB:CC:DD:EE:FF
```


```
docker run -d --restart=unless-stopped --network=host \
--privileged -v /var/run/dbus/:/var/run/dbus/:z lywsd03mmc-prom-expo \
--listen-address 127.0.0.1:8091 --bt-scan-timeout 31s \
--dev 00:11:22:33:44:55 --dev AA:BB:CC:DD:EE:FF
```

## Example output 
`curl http://127.0.0.1:8081/metrics`
```
HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP sensor_battery_volts Battery voltage
# TYPE sensor_battery_volts gauge
sensor_battery_volts{mac="A4:C1:38:8A:3B:DE"} 2.883 1744797845657
sensor_battery_volts{mac="A4:C1:38:B4:96:0B"} 2.801 1744797844440
# HELP sensor_humidity_percent Humidity in percent
# TYPE sensor_humidity_percent gauge
sensor_humidity_percent{mac="A4:C1:38:8A:3B:DE"} 57.09 1744797845657
sensor_humidity_percent{mac="A4:C1:38:B4:96:0B"} 40.98 1744797844440
# HELP sensor_temp_celsius Temperature in celsius
# TYPE sensor_temp_celsius gauge
sensor_temp_celsius{mac="A4:C1:38:8A:3B:DE"} 5.69 1744797845657
sensor_temp_celsius{mac="A4:C1:38:B4:96:0B"} 22.85 1744797844440
```


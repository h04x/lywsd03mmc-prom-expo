# LYWSD03MMC prometheus exporter

Currently only supported [pvvx/ATC_MiThermometer](https://github.com/pvvx/ATC_MiThermometer) firmware. 
Custom advertisement format

## Requirements 

`Bluez` used on linux box, it must be installed on host machine

## Building

`make build`

## Usage
See all args `lywsd03mmc-prom-expo -h`  
Minimal run, pass thermometers
```
./lywsd03mmc-prom-expo --dev AA:BB:CC:DD:EE:FF --dev 11:22:33:44:55:66
```

Now you can request metrics `curl http://127.0.0.1:8080/metrics`, the response  will look something like this
```
# HELP sensor_battery_volts Battery voltage
# TYPE sensor_battery_volts gauge
sensor_battery_volts{mac="AA:BB:CC:DD:EE:FF"} 2.883 1744797845657
sensor_battery_volts{mac="11:22:33:44:55:66"} 2.801 1744797844440
# HELP sensor_humidity_percent Humidity in percent
# TYPE sensor_humidity_percent gauge
sensor_humidity_percent{mac="AA:BB:CC:DD:EE:FF"} 57.09 1744797845657
sensor_humidity_percent{mac="11:22:33:44:55:66"} 40.98 1744797844440
# HELP sensor_temp_celsius Temperature in celsius
# TYPE sensor_temp_celsius gauge
sensor_temp_celsius{mac="AA:BB:CC:DD:EE:FF"} 5.69 1744797845657
sensor_temp_celsius{mac="11:22:33:44:55:66"} 22.85 1744797844440
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
Build image
```
docker buildx build -t lywsd03mmc-prom-expo .
```
Debug run example
```
docker run -it --rm --network=host --privileged \
-v /var/run/dbus/:/var/run/dbus/:z lywsd03mmc-prom-expo \
--listen-address 127.0.0.1:8081 --bt-scan-timeout 31s \
--dev 00:11:22:33:44:55 --dev AA:BB:CC:DD:EE:FF
```

Prod run example
```
docker run -d --restart=unless-stopped --network=host \
--privileged -v /var/run/dbus/:/var/run/dbus/:z lywsd03mmc-prom-expo \
--listen-address 127.0.0.1:8081 --bt-scan-timeout 31s \
--dev 00:11:22:33:44:55 --dev AA:BB:CC:DD:EE:FF
```

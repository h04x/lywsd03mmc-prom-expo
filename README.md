# LYWSD03MMC prometheus exporter

LYWSD03MMC temp, humidity, voltage prometheus exporter  
Use it for https://github.com/pvvx/ATC_MiThermometer firmware. Custom advertisement format  


## Installation systemd
Build and copy config files, run  
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


## Running for testing
For playing around args withount installation
```
make run ARG="--dev 11:22:33:44:55:66 --dev AA:BB:CC:DD:EE:FF"
```


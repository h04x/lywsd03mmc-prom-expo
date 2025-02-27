package main

import (
	"fmt"
	"lywsd03mmc-prom-expo/poller"
	"time"
)

const scanTimeoutSec = 10

func main() {

	p := poller.NewDevicePoller("A4:C1:38:8A:3B:DE", scanTimeoutSec)
	for {
		temp, humidity, voltage, err := p.Poll()
		if err != nil {
			fmt.Println(err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		fmt.Printf("%v %v°C %v%% %vV\n", time.Now().Format("15:04"), temp, humidity, voltage)
		time.Sleep(time.Minute * 10)
	}

}

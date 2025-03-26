build:
	go build

run: build
	./lywsd03mmc-prom-expo $(ARG)

install: build
	cp lywsd03mmc-prom-expo /usr/local/bin/
	cp lywsd03mmc-prom-expo.service /etc/systemd/system/
	cp lywsd03mmc-prom-expo.env /etc/default/lywsd03mmc-prom-expo

clean:
	go clean

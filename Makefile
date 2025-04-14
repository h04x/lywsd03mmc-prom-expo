build:
	go build -o lywsd03mmc-prom-expo main.go

run: build
	./lywsd03mmc-prom-expo $(ARG)

install: build
	cp lywsd03mmc-prom-expo /usr/local/bin/
	cp --update lywsd03mmc-prom-expo.service /etc/systemd/system/
	cp --update lywsd03mmc-prom-expo.env /etc/default/lywsd03mmc-prom-expo

clean:
	go clean

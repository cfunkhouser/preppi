build:
	mkdir -p bin
	go build -i -v -ldflags="-X main.buildID=$(shell git describe --always --long --dirty)" -o bin/preppi main.go

clean:
	rm -rf bin

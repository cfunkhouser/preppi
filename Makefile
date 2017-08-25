all: outdir build

outdir:
	mkdir -p bin

dockerprep:
	docker build -t preppi-build build/

build: outdir
	go build -i -v -ldflags="-X main.buildID=$(shell git describe --always --long --dirty)" -o bin/preppi main.go

linux: outdir
	docker run -ti --rm -v $(shell pwd):/opt/gopath/src/github.com/cfunkhouser/preppi preppi-build go build -i -v -ldflags="-X main.buildID=$(shell git describe --always --long --dirty)" -o bin/preppi main.go

clean:
	rm -rf bin

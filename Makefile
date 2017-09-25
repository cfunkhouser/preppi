SRCROOT = /opt/gopath/src/github.com/cfunkhouser/preppi

VERSION = 0.1.1
BUILDID = $(shell git describe --always --long --dirty)

LDFLAGS="-X github.com/cfunkhouser/preppi/preppi.Version=$(VERSION) -X github.com/cfunkhouser/preppi/preppi.BuildID=$(BUILDID)"

all: build

prep:
	mkdir -p build/out/bin
	echo "$(VERSION)-$(BUILDID)" > build/out/VERSION

docker:
	docker build -t preppi-build build/

linux-i386: prep
	env GOOS=linux GOARCH=386 go build -v -ldflags=$(LDFLAGS) -o build/out/bin/preppi-linux-i386 main.go

linux-armhf: prep
	env GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags=$(LDFLAGS) -o build/out/bin/preppi-linux-armhf main.go

linux-amd64: prep
	env GOOS=linux GOARCH=amd64 go build -v -ldflags=$(LDFLAGS) -o build/out/bin/preppi-linux-amd64 main.go

linux: linux-amd64 linux-armhf linux-i386

build: prep
	go build -i -v -ldflags=$(LDFLAGS) -o build/out/bin/preppi main.go

deb-i386: docker linux-i386
	docker run -ti --rm -v $(shell pwd):$(SRCROOT) preppi-build build-preppi-deb.sh i386

deb-amd64: docker linux-amd64
	docker run -ti --rm -v $(shell pwd):$(SRCROOT) preppi-build build-preppi-deb.sh amd64

deb-armhf: docker linux-armhf
	docker run -ti --rm -v $(shell pwd):$(SRCROOT) preppi-build build-preppi-deb.sh armhf

deb: docker linux
	docker run -ti --rm -v $(shell pwd):$(SRCROOT) preppi-build build-preppi-deb.sh

clean:
	rm -rf build/out

superclean: clean
	docker rmi -f preppi-build

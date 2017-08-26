SRCROOT = /opt/gopath/src/github.com/cfunkhouser/preppi

VERSION = 0.1
BUILDID = $(shell git describe --always --long --dirty)

all: build

prep:
	mkdir -p build/out/bin
	echo "$(VERSION)-$(BUILDID)" > build/out/VERSION

docker:
	docker build -t preppi-build build/

build: prep
	go build -i -v -ldflags="-X main.buildID=$(BUILDID) -X main.version=$(VERSION)" -o build/out/bin/preppi main.go

linux: docker prep
	docker run -ti --rm -v $(shell pwd):$(SRCROOT) preppi-build build-preppi.sh

deb: linux
	docker run -ti --rm -v $(shell pwd):$(SRCROOT) preppi-build build-preppi-deb.sh

clean:
	rm -rf build/out

superclean: clean
	docker rmi -f preppi-build

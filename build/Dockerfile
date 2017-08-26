FROM debian:jessie-slim
MAINTAINER Christian Funkhouser <christian.funkhouser@gmail.com>
ENV DEBIAN_FRONTEND noninteractive

ENV GOPATH "/opt/gopath"
ENV GOTARBALL "https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz"
ENV PATH "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:${GOPATH}/bin"

# Install dependencies for the following commands
RUN apt-get update -q -y --fix-missing && \
    apt-get install -q -y --no-install-recommends \
      build-essential \
      ca-certificates \
      curl \
      git-core \
      sudo

# Make sure GOPATH exists
RUN mkdir -p "${GOPATH}" 

# Install the Go binaries under /usr/local.
RUN curl -s "${GOTARBALL}" | sudo tar -zx -C "/usr/local/"

# Install helpful tools for building Go code
RUN go get github.com/golang/dep/cmd/dep

# Copy the build scripts, and make sure they're executable.
ADD build-preppi.sh /usr/local/bin/build-preppi.sh
ADD build-preppi-deb.sh /usr/local/bin/build-preppi-deb.sh
RUN chmod +x /usr/local/bin/build-preppi.sh && \
    chmod +x /usr/local/bin/build-preppi-deb.sh

CMD bash
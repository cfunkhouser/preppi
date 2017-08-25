FROM debian:jessie-slim
MAINTAINER Christian Funkhouser <christian.funkhouser@gmail.com>
ENV DEBIAN_FRONTEND noninteractive

ENV GOPATH "/opt/gopath"
ENV GOTARBALL "https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz"
ENV PATH "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:${GOPATH}/bin"

# Install dependencies for the following commands
RUN apt-get update -q -y --fix-missing && \
    apt-get install -q -y --no-install-recommends \
      apt-transport-https \
      build-essential \
      ca-certificates \
      curl \
      file \
      git-core \
      gnupg2 \
      software-properties-common \
      sudo \
    && \
    # Create an empty Go workspace at $GOPATH
    mkdir -p "${GOPATH}/{bin,pkg,src}"

# Install the Go binaries under /usr/local.
RUN /usr/bin/curl -s "${GOTARBALL}" | sudo tar -zx -C "/usr/local/"

# Install helpful tools for building Go code
RUN /usr/local/go/bin/go get github.com/golang/dep/cmd/dep && \
    /usr/local/go/bin/go get github.com/mitchellh/gox

CMD /bin/bash
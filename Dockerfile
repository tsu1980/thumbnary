# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM asia.gcr.io/chocorail-1919/thumbnary-buildbase:8.7.0 as builder
MAINTAINER tsu1980@gmail.com

# Fetch the latest version of the package
RUN go get -u golang.org/x/net/context
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get -u github.com/cenkalti/backoff
RUN go get -u github.com/gomodule/redigo/redis
RUN go get -u github.com/spf13/viper

# Copy thumbnary sources
COPY . $GOPATH/src/github.com/tsu1980/thumbnary

# Compile thumbnary
RUN go build -o bin/thumbnary github.com/tsu1980/thumbnary

FROM ubuntu:16.04

RUN \
  # Install runtime dependencies
  apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
  libglib2.0-0 libjpeg-turbo8 libpng12-0 libopenexr22 \
  libwebp5 libtiff5 libgif7 libexif12 libxml2 libpoppler-glib8 \
  libmagickwand-6.q16-2 libpango1.0-0 libmatio2 libopenslide0 \
  libgsf-1-114 fftw3 liborc-0.4 librsvg2-2 libcfitsio2 && \
  # Clean up
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /usr/local/lib /usr/local/lib
RUN ldconfig
COPY --from=builder /go/bin/thumbnary bin/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

# Server port to listen
ENV PORT 9000

# Run the entrypoint command by default when the container starts.
ENTRYPOINT ["bin/thumbnary"]

# Expose the server TCP port
EXPOSE 9000

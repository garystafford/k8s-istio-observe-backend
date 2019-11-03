# Build based tips from:
# https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324

################################
# STEP 1 build executable binary
################################
FROM golang:1.13-alpine AS builder

# Install git, zoneinfo, and SSL certs
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Create unprivileged appuser
RUN adduser -D -g '' appuser

# Copy file(s)
WORKDIR /go/src/app
COPY main.go .

# Using go get
RUN go get -d -v

# Disable crosscompiling
ENV CGO_ENABLED=0

# Compile Linux only
ENV GOOS=linux

# Build the binary - remove debug info and compile only for linux target
RUN go build  -ldflags '-w -s' -a -installsuffix cgo -o /go/bin/app .

############################
# STEP 2 build a small image
############################
FROM scratch

# Import the user and group files from the builder
COPY --from=builder /etc/passwd /etc/passwd

# Import the zoneinfo and SSL cert files from the builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy static executable
COPY --from=builder /go/bin/app /go/bin/app

# Use an unprivileged user
USER appuser

# Run the app binary
ENTRYPOINT ["/go/bin/app"]
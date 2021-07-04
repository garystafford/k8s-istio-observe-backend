# build based on:
# https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
# date: 2021-07-03

################################
# STEP 1 build executable binary
################################
FROM golang:1.16.5-alpine3.14 AS builder

# Use port 50051 for gRPC, 8080 for http
EXPOSE 50051

# Install git, zoneinfo, and SSL certs
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Create unprivileged appuser
RUN adduser -D -g '' appuser

# Copy file(s)
WORKDIR /go/src/app
COPY main.go .

# creates go.mod file to track code dependencies
RUN go mod init

# ensures go.mod file matches source code in the module
RUN go mod tidy

# Disable crosscompiling
ENV CGO_ENABLED=0

# Compile Linux only
ENV GOOS=linux

# Build the binary - remove debug info and compile only for linux target
RUN go build -ldflags '-w -s' -a -installsuffix cgo -o /go/bin/app .

############################
# STEP 2 build a small image
############################
FROM scratch

LABEL maintainer="Gary A. Stafford <gary.a.stafford@gmail.com>"
ENV REFRESHED_AT 2021-07-03

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

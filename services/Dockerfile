# https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324

############################
# STEP 1 build executable binary
############################
FROM golang:1.12-alpine AS builder

# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
# Create appuser.
# RUN adduser -D -g '' appuser
# Copy files
WORKDIR /go/src/hello
COPY main.go .
# Using go get.
RUN go get -d -v
# Using go mod.
RUN go mod download
# Build the binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/hello .
# RUN go build -o /go/bin/app

############################
# STEP 2 build a small image
############################
FROM alpine:3.9

RUN apk update && apk add --no-cache ca-certificates

# COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable.
COPY --from=builder /go/bin/hello /bin/hello
# Use an unprivileged user.
# USER appuser
# Run the hello binary.
ENTRYPOINT ["/bin/hello"]
############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
# Create appuser.
#RUN adduser -D -g '' appuser

WORKDIR /go/src/app
COPY main.go .

# Using go get.
RUN go get -d -v

# Build the binary.
# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/app .
RUN go build -o /go/bin/app

############################
# STEP 2 build a small image
############################
FROM alpine:latest

# Import the user and group files from the builder.
#COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable.
COPY --from=builder /go/bin/app /go/bin/app
# Use an unprivileged user.
#USER appuser
# Run the hello binary.
ENTRYPOINT ["/go/bin/app"]
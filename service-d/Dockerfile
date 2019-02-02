FROM golang:onbuild
LABEL maintainer="Gary A. Stafford <garystafford@rochester.rr.com>"
ENV REFRESHED_AT 2019-02-02
EXPOSE 8000

RUN mkdir /app
ADD main.go /app/main.go
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]

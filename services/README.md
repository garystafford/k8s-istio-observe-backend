# Notes

Build all services locally:

```shell
go version
# go version go1.16.3 darwin/amd64

sh ./part0_build_servcies_locally.sh
```

Run a service locally:

```shell
cd ./service/service-a
go mod tidy
go run *.go
```

To test service-a, from a separate terminal window:

```shell
http http://localhost:80/api/ping
```

Build all Docker images:

```shell
cd services/
time | sh ./part1_build_srv_images.sh
```

Push all Docker images.

```shell
sh time | ./part2_push_images.sh
```

```shell
time | sh ./part1_build_srv_images.sh && sh ./part2_push_images.sh
```

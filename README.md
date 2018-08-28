# prometheus_docker_exporter




### build

```shell

$ go build -o prometheus_docker_exporter main.go  collectors.go

$ GOOS=linux GOARCH=amd64 go build -o prometheus_docker_exporter_linux main.go collectors.go

$ docker build -t cwr0401/prometheus_docker_exporter:latest .

```

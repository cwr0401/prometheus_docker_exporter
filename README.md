# prometheus_docker_exporter

prometheus docker exporter 用于使用 prometheus 抓取 docker containers stats 信息。


### build

```shell
$ go build -o prometheus_docker_exporter main.go metrics.go collectors.go service.go 

$ GOOS=linux GOARCH=amd64 go build -o prometheus_docker_exporter_linux main.go metrics.go collectors.go service.go

$ docker build -t cwr0401/prometheus_docker_exporter:latest .

```


### run 
```shell
$ docker pull cwr0401/prometheus_docker_exporter:latest

$ docker run -it -d --rm \
-p 8000:8000  \
-e TZ="Asia/Shanghai"  \
-e CONSUL_SERVICE_NAME="test-prom-docker-metrics"  \
-e CONSUL_SERVICE_ID="test-prom-docker-metrics-01" \
-e CONSUL_SERVICE_PORT=8000 \
-v /var/run/docker.sock:/var/run/docker.sock:ro
cwr0401/prometheus_docker_exporte

$ curl http://127.0.0.1:8000/health
ok

$ curl http://127.0.0.1:8000/metrics
```
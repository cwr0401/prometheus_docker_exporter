# prometheus_docker_exporter



### build

```shell
$ go build -o prometheus_docker_exporter main.go collectors.go service.go metrics.go

$ GOOS=linux GOARCH=amd64 go build -o prometheus_docker_exporter_linux main.go collectors.go service.go

$ docker build -t cwr0401/prometheus_docker_exporter:latest .

```

### 容器环境变量
TZ Asia/Shanghai    时区

PROMETHEUS_SCRAPE_PORT   

PROMETHEUS_SCRAPE_IP

consul 服务注册信息

CONSUL_SERVICE_ID

CONSUL_ADDRESS

CONSUL_DATACENTER

CONSUL_SERVICE_NAME

CONSUL_SERVICE_TAG

package main

import (
	"os"
	"math/rand"
	"time"
	"strconv"
	"net"
	"github.com/hashicorp/consul/api"
    log "github.com/sirupsen/logrus"
	"fmt"
	"runtime"
)

func registerConsulService() {
	// CONSUL_ADDRESS
	// CONSUL_DATACENTER
	// CONSUL_USERNAME
	// CONSUL_PASSWORD
	// CONSUL_TOKEN
	// CONSUL_SERVICE_ID_SUFFIX
	// CONSUL_SERVICE_NAME

	consulConfig := api.DefaultConfig()

	// setting consul
	consulAddress := os.Getenv("CONSUL_ADDRESS")
	if consulAddress == "" {
		log.Warn("CONSUL_ADDRESS is not setting, use default (127.0.0.1:8500).")
		consulConfig.Address = "127.0.0.1:8500"
	} else {
		consulConfig.Address = consulAddress
	}

	consulDatacenter := os.Getenv("CONSUL_DATACENTER")
	if consulDatacenter == "" {
		log.Warn("CONSUL_DATACENTER is not setting, use default (prometheus)")
		consulConfig.Datacenter = "prometheus"
	} else {
		consulConfig.Datacenter = consulDatacenter
	}

	consulToken := os.Getenv("CONSUL_TOKEN")
	if consulToken != "" {
		consulConfig.Token = consulToken
	}

	consulUsername := os.Getenv("CONSUL_USERNAME")
	consulPassword := os.Getenv("CONSUL_PASSWORD")
	if consulUsername != "" && consulPassword != "" {
		consulHttpAuth := &api.HttpBasicAuth{
			Username: consulUsername,
			Password: consulPassword,
		}
		consulConfig.HttpAuth = consulHttpAuth
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.Panic("Init Consul client error: ", err)
	}

	consulServiceName := os.Getenv("CONSUL_SERVICE_NAME")
	if consulServiceName == "" {
		log.Warn("CONSUL_SERVICE_NAME is not setting, use default (prometheus-docker-metrics)")
		consulServiceName = "prometheus-docker-metrics"
	}
	consulServiceID := os.Getenv("CONSUL_SERVICE_ID")
	if consulServiceID == "" {
		consulServiceID = "prometheus-docker-metrics-" + consulServiceIDSuffix()
		log.Warn("CONSUL_SERVICE_ID is not setting, use default ", consulServiceID)
	}
	serviceIP := os.Getenv("PROMETHEUS_SCRAPE_IP")
	if serviceIP == "" {
		// use eth0 network interface ipv4 address
		interfaceName := "eth0"
		// mac os
		if runtime.GOOS == "darwin" {
			interfaceName = "en0"
		}
		serviceIP = getInterfaceIP(interfaceName)
		log.Warn("PROMETHEUS_SCRAPE_IP is not setting, use container ip ", serviceIP)
	}
	servicePort, err := strconv.Atoi(os.Getenv("PROMETHEUS_SCRAPE_PORT"))
	if err != nil {
		log.Warn("PROMETHEUS_SCRAPE_PORT is not setting, use default (8765)")
		servicePort = 8765
	}

	registration := new(api.AgentServiceRegistration)
	registration.ID = consulServiceID
	registration.Name = consulServiceName

	registration.Address = serviceIP
	registration.Port = servicePort
	consulServiceTag := os.Getenv("CONSUL_SERVICE_TAG")
	if consulServiceTag != "" {
		registration.Tags = []string{consulServiceTag}
	}

	check := new(api.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("http://%s:%d/%s", registration.Address, registration.Port, "health")
	check.Timeout = "3s"
	check.Interval = "30s"
	check.Method = "GET"
	registration.Check = check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Panic("Register consul service error: ", err)
	}
	log.Infof("Register consul service: %s, %s, %s, %d",
		registration.Name,
		registration.ID,
		registration.Address,
		registration.Port)
}

func getInterfaceIP(interfaceName string) string {
	inter, err := net.InterfaceByName(interfaceName)
	if err != nil {
		log.Error("Get network interface ", interfaceName, " error: ", err)
		return "127.0.0.1"
	}
	interAddresses, err := inter.Addrs()
	if err != nil {
		log.Error("Get network interface ", interfaceName, " address error", err)
		return "127.0.0.1"
	}
	for _, interAddr := range interAddresses {
		ip, _, err := net.ParseCIDR(interAddr.String())
		if err != nil{
			log.Warn("Parse CIDR error: ", err)
			continue
		}
		ipv4 := ip.To4()
		if ipv4 != nil {
			return ipv4.String()
		}
	}
	return "127.0.0.1"
}

// 生成随机 Service ID 后缀
func consulServiceIDSuffix() string{
	var suffix string
	rand.Seed(time.Now().UnixNano())
	hexNum := [16]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	for i := 0; i < 5; i++ {
		suffix += hexNum[rand.Intn(16)]
	}
	return suffix
}
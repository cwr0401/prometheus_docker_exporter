package main

import (
	"math/rand"
	"time"

	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-sockaddr/template"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"runtime"
)

func before(c *cli.Context) error {
	// debug level if requested by user
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	consulConfig := api.DefaultConfig()

	// setting consul
	consulConfig.Address = c.String("consul-address")
	consulConfig.Datacenter = c.String("consul-dc")

	consulToken := c.String("consul-token")
	if consulToken != "" {
		consulConfig.Token = consulToken
	}

	consulUsername := c.String("consul-username")
	consulPassword := c.String("consul-password")
	if consulUsername != "" && consulPassword != "" {
		consulHttpAuth := &api.HttpBasicAuth{
			Username: consulUsername,
			Password: consulPassword,
		}
		consulConfig.HttpAuth = consulHttpAuth
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.Error("Init Consul client error: ", err)
		return err
	}

	// service register info
	registration := new(api.AgentServiceRegistration)

	registration.Name = c.String("service-name")

	consulServiceID := c.String("service-id")
	if consulServiceID == "" {
		consulServiceID = "prometheus-docker-metrics-" + consulServiceIDSuffix()
		log.Warn("Consul service id is not setting, use default ", consulServiceID)
	}
	registration.ID = consulServiceID

	serviceIP := c.String("service-ip")
	if serviceIP == "" {
		// use eth0 network interface ipv4 address
		// mac os
		if runtime.GOOS == "darwin" {
			serviceIP, err = template.Parse(`{{ GetInterfaceIP "en0" }}`)
		} else {
			serviceIP, err = template.Parse(`{{ GetInterfaceIP "eth0" }}`)
		}
		if err != nil {
			serviceIP = "127.0.0.1"
		}
		log.Warn("Consul service ip is not setting, use container ip ", serviceIP)
	}
	registration.Address = serviceIP

	servicePort := c.Uint("service-port")
	if servicePort == 0 || servicePort > 65535 {
		log.Warn("Consul service id invalid, use default (8765)")
		servicePort = 8765
	}
	registration.Port = int(servicePort)

	consulServiceTag := c.StringSlice("service-tags")
	if consulServiceTag != nil {
		registration.Tags = consulServiceTag
	}

	check := new(api.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("http://%s:%d/%s", registration.Address, registration.Port, "health")
	check.Timeout = "3s"
	check.Interval = "30s"
	check.Method = "GET"
	registration.Check = check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Error("Register Consul service error: ", err)
		return err
	}
	log.Infof("Register consul service: %s, %s, %s, %d",
		registration.Name,
		registration.ID,
		registration.Address,
		registration.Port)

	return nil
}

//func getInterfaceIP(interfaceName string) string {
//	inter, err := net.InterfaceByName(interfaceName)
//	if err != nil {
//		log.Error("Get network interface ", interfaceName, " error: ", err)
//		return "127.0.0.1"
//	}
//	interAddresses, err := inter.Addrs()
//	if err != nil {
//		log.Error("Get network interface ", interfaceName, " address error", err)
//		return "127.0.0.1"
//	}
//	for _, interAddr := range interAddresses {
//		ip, _, err := net.ParseCIDR(interAddr.String())
//		if err != nil {
//			log.Warn("Parse CIDR error: ", err)
//			continue
//		}
//		ipv4 := ip.To4()
//		if ipv4 != nil {
//			return ipv4.String()
//		}
//	}
//	return "127.0.0.1"
//}

// 生成随机 Service ID 后缀
func consulServiceIDSuffix() string {
	var suffix string
	rand.Seed(time.Now().UnixNano())
	hexNum := [16]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	for i := 0; i < 5; i++ {
		suffix += hexNum[rand.Intn(16)]
	}
	return suffix
}

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "DEBUG_MODE",
		Name:   "debug",
		Usage:  "enable app debug mode",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_ADDRESS",
		Name:   "consul-address",
		Usage:  "address of the Consul server",
		Value:  "127.0.0.1:8500",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_DATACENTER",
		Name:   "consul-dc",
		Usage:  "Consul datacenter to use",
		Value:  "prometheus",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_TOKEN",
		Name:   "consul-token",
		Usage:  "Consul token is used to provide a per-request ACL token which overrides the agent's default token",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_USERNAME",
		Name:   "consul-username",
		Usage:  "Consul username for http access in httpAuth mode",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_PASSWORD",
		Name:   "consul-password",
		Usage:  "Consul password for http access in httpAuth mode",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_SERVICE_NAME",
		Name:   "service-name",
		Value:  "prometheus-docker-metrics",
		Usage:  "service name register to Consul",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_SERVICE_ID",
		Name:   "service-id",
		Usage:  "service ID register to Consul",
	},
	cli.StringFlag{
		EnvVar: "CONSUL_SERVICE_IP",
		Name:   "service-ip",
		Usage:  "service ip register to Consul",
	},
	cli.UintFlag{
		EnvVar: "CONSUL_SERVICE_PORT",
		Name:   "service-port",
		Usage:  "service port register to Consul",
		Value:  8765,
	},
	cli.StringSliceFlag{
		EnvVar: "CONSUL_SERVICE_TAG",
		Name:   "service-tags",
		Usage:  "service tag register to Consul",
	},
	cli.StringFlag{
		EnvVar: "SERVER_ADDR",
		Name:   "server-addr",
		Usage:  "server address",
		Value:  ":8000",
	},
}

func metricServer(c *cli.Context) error {
	// docker client
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Error("", err)
		return err
	}

	// health check
	http.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain")
		writer.Header().Set("Cache-Control", "no-cache,no-store")
		writer.Header().Set("Server", "prometheus")
		writer.Write([]byte("ok\n"))
	})

	http.Handle("/", handler)
	http.Handle("/metrics", handler)

	go func() {
		err = http.ListenAndServe(c.String("server-addr"), nil)
		if err != nil {
			log.Panic("HTTP Server Listen failed.")
		}
	}()

	for true {
		log.Info("Get Containers stats.")
		containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			log.Error("Get container list error: ", err)
			return err
		}
		scrapeNumber.Inc()
		statsNumber.Set(0)
		for _, container := range containers {
			go containerToMetrics(client, container)
		}
		time.Sleep(time.Minute)
	}

	return nil
}

func containerToMetrics(client *client.Client, container types.Container) error {
	name := container.Names[0][1:]
	shortID := container.ID[:10]
	log.Infof("Container Name %v (ID: %s)", name, shortID)
	resp, err := client.ContainerStats(context.Background(), container.ID, false)
	if err != nil {
		log.Errorf("Container Name %v (ID: %s) get container stats error: %s", name, shortID, err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Container Name: %v (ID: %s) read container stats data error: ", name, shortID, err)
		return nil
	}

	var containerStats types.StatsJSON
	if err = json.Unmarshal(body, &containerStats); err != nil {
		log.Errorf("Container Name: %v (ID: %s) format container stats data to json error: %s", name, shortID, err)
		return nil
	}
	if container.ID != containerStats.ID {
		log.Error("Container ID Inconsistent.")
		return nil
	}

	statsNumber.Inc()
	containerName := containerStats.Name[1:]

	memoryLimit.WithLabelValues(containerName, shortID).Set(float64(containerStats.MemoryStats.Limit))
	memoryUsage.WithLabelValues(containerName, shortID).Set(float64(containerStats.MemoryStats.Usage))
	rss, ok := containerStats.MemoryStats.Stats["rss"]
	if ok {
		memoryRss.WithLabelValues(containerName, shortID).Set(float64(rss))
	} else {
		log.Warnf("Container Name %v (ID: %s) stats not rss field", name, shortID)
	}

	cpuUser.WithLabelValues(containerName, shortID).Set(float64(containerStats.CPUStats.CPUUsage.UsageInUsermode))
	cpuKernel.WithLabelValues(containerName, shortID).Set(float64(containerStats.CPUStats.CPUUsage.UsageInKernelmode))
	cpuAll.WithLabelValues(containerName, shortID).Set(float64(containerStats.CPUStats.CPUUsage.TotalUsage))
	cpuSystem.WithLabelValues(containerName, shortID).Set(float64(containerStats.CPUStats.SystemUsage))

	for netName, network := range containerStats.Networks {
		rxBytes.WithLabelValues(containerName, shortID, netName).Set(float64(network.RxBytes))
		rxPackets.WithLabelValues(containerName, shortID, netName).Set(float64(network.RxPackets))
		txBytes.WithLabelValues(containerName, shortID, netName).Set(float64(network.TxBytes))
		txPackets.WithLabelValues(containerName, shortID, netName).Set(float64(network.TxPackets))
	}

	return nil
}

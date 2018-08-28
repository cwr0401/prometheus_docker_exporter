package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Panic(err)
	}

	http.Handle("/metrics", handler)
	go func() {
		err = http.ListenAndServe(":8000", nil)
		if err != nil {
			log.Panic("HTTP Server Listen failed.")
		}
	}()

	for true {
		log.Info("Get Containers stats.")
		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			log.Panic(err)
		}
		scrapeNumber.Inc()
		statsNumber.Set(0)
		for _, container := range containers {
			go containerToMetrics(cli, container)
		}
		time.Sleep(time.Minute)
	}
}

func containerToMetrics(cli *client.Client, container types.Container) error {
	name := container.Names[0][1:]
	shortID := container.ID[:10]
	log.Infof("Container Name %v (ID: %s)", name, shortID)
	resp, err := cli.ContainerStats(context.Background(), container.ID, false)
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

	var container_stats types.StatsJSON
	if err = json.Unmarshal(body, &container_stats); err != nil {
		log.Errorf("Container Name: %v (ID: %s) format container stats data to json error: %s", name, shortID, err)
		return nil
	}
	if container.ID != container_stats.ID {
		log.Error("Container ID Inconsistent.")
		return nil
	}

	statsNumber.Inc()
	contaner_name := container_stats.Name[1:]

	memoryLimit.WithLabelValues(contaner_name, shortID).Set(float64(container_stats.MemoryStats.Limit))
	memoryUsage.WithLabelValues(contaner_name, shortID).Set(float64(container_stats.MemoryStats.Usage))
	rss, ok := container_stats.MemoryStats.Stats["rss"]
	if ok {
		memoryRss.WithLabelValues(contaner_name, shortID).Set(float64(rss))
	} else {
		log.Warnf("Container Name %v (ID: %s) stats not rss field", name, shortID)
	}

	cpuUser.WithLabelValues(contaner_name, shortID).Set(float64(container_stats.CPUStats.CPUUsage.UsageInUsermode))
	cpuKernel.WithLabelValues(contaner_name, shortID).Set(float64(container_stats.CPUStats.CPUUsage.UsageInKernelmode))
	cpuAll.WithLabelValues(contaner_name, shortID).Set(float64(container_stats.CPUStats.CPUUsage.TotalUsage))
	cpuSystem.WithLabelValues(contaner_name, shortID).Set(float64(container_stats.CPUStats.SystemUsage))

	for netName, network := range container_stats.Networks {
		rxBytes.WithLabelValues(contaner_name, shortID, netName).Set(float64(network.RxBytes))
		rxPackets.WithLabelValues(contaner_name, shortID, netName).Set(float64(network.RxPackets))
		txBytes.WithLabelValues(contaner_name, shortID, netName).Set(float64(network.TxBytes))
		txPackets.WithLabelValues(contaner_name, shortID, netName).Set(float64(network.TxPackets))
	}

	return nil
}

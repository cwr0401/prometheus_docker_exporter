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
)

// health check handler
func healthHandler(w http.ResponseWriter, reg *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-cache,no-store")
	w.Header().Set("Server", "prometheus")
	w.Write([]byte("ok\n"))
}

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Panic(err)
	}

	// health check
	http.HandleFunc("/health", healthHandler)

	http.Handle("/metrics", handler)

	go func() {
		err = http.ListenAndServe(":8000", nil)
		if err != nil {
			log.Panic("HTTP Server Listen failed.")
		}
	}()

	go registerConsulService()

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

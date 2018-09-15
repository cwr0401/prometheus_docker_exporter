package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	memoryLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_memory_stats_limit",
		Help: "Memory Limit.",
	},
		[]string{"container_name", "container_id"})
	memoryUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_memory_stats_usage",
		Help: "Total memory usage, include Virtual Memory Size.",
	},
		[]string{"container_name", "container_id"})
	memoryRss = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_memory_stats_rss",
		Help: "Resident Memory Size.",
	},
		[]string{"container_name", "container_id"})
	cpuUser = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_cpu_stats_usermode",
		Help: "time running un-niced user processes.",
	},
		[]string{"container_name", "container_id"})
	cpuKernel = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_cpu_stats_kernelmode",
		Help: "time running kernel processes.",
	},
		[]string{"container_name", "container_id"})
	cpuAll = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_cpu_stats_all",
		Help: "total cpu time for container.",
	},
		[]string{"container_name", "container_id"})
	cpuSystem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_cpu_stats_system",
		Help: "host total cpu time.",
	},
		[]string{"container_name", "container_id"})
	rxBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_networks_rx_bytes",
		Help: "network received bytes.",
	},
		[]string{"container_name", "container_id", "interface"})
	rxPackets = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_networks_rx_packets",
		Help: "network received packets.",
	},
		[]string{"container_name", "container_id", "interface"})
	txBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_networks_tx_bytes",
		Help: "network send bytes.",
	},
		[]string{"container_name", "container_id", "interface"})
	txPackets = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_container_networks_tx_packets",
		Help: "network send packets.",
	},
		[]string{"container_name", "container_id", "interface"})
	scrapeNumber = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "docker_container_scrape_total",
		Help: "the number of scrape."})
	statsNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "docker_container_stats_num",
		Help: "The amount of docker container stats"})
	registry                     = prometheus.NewRegistry()
	gather   prometheus.Gatherer = registry
	handler                      = promhttp.HandlerFor(gather, promhttp.HandlerOpts{})
)

func init() {
	// log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	// log.SetLevel(log.DebugLevel)

	// Metrics have to be registered to be exposed:
	registry.MustRegister(memoryLimit)
	registry.MustRegister(memoryUsage)
	registry.MustRegister(memoryRss)
	registry.MustRegister(cpuUser)
	registry.MustRegister(cpuKernel)
	registry.MustRegister(cpuAll)
	registry.MustRegister(cpuSystem)
	registry.MustRegister(rxBytes)
	registry.MustRegister(rxPackets)
	registry.MustRegister(txBytes)
	registry.MustRegister(txPackets)
	registry.MustRegister(scrapeNumber)
	registry.MustRegister(statsNumber)
}

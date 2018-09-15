package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "prometheus_docker_exporter"
	app.Version = "0.1.8"
	app.Usage = "docker metrics server"
	app.Flags = flags
	app.Before = before
	app.Action = metricServer

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

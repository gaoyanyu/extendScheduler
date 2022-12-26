package main

import (
	"extendScheduler/pkg/disk"
	"fmt"
	"math/rand"
	"os"
	"time"

	"k8s.io/kubernetes/cmd/kube-scheduler/app"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	command := app.NewSchedulerCommand(
		app.WithPlugin(disk.Name, disk.New))

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

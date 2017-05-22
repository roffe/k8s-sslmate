package main

import (
	"github.com/robfig/cron"
	"k8s.io/client-go/kubernetes"
)

func startCron(clientset *kubernetes.Clientset) {
	c := cron.New()
	c.AddFunc("@every 60m", func() { run_SSLmate() })
	c.AddFunc("@every 1m", func() { getConfigmap(clientset, false) })
	c.Start()

}

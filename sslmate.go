package main

import (
	"log"
	"os/exec"
)

func run_SSLmate() bool {
	cmd := exec.Command("sslmate", "download", "--all")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("FATAL: %s", err.Error())
	}
	log.Printf("INFO: Waiting for SSLmate to finish...")
	err = cmd.Wait()
	if err != nil {
		log.Print("INFO: SSLmate had no new certs to download")
	}
	if err == nil {
		log.Print("INFO: SSLmate has downloaded new certs")
	}
	return true
}

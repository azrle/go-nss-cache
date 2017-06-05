package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/azrle/go-nss-cache/client"
	"github.com/golang/glog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Source         string
	UpdateInterval string
	GCEMetadata    client.GCEMetadataConf

	UpdateIntervalDuration time.Duration
}

func main() {
	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	configPath := flag.String("config", "/etc/go-nss-cache.toml", "Toml config file path")
	flag.Parse()

	var config Config
	var err error
	if _, err = toml.DecodeFile(*configPath, &config); err != nil {
		glog.Fatal(err)
		return
	}
	if config.UpdateIntervalDuration, err = time.ParseDuration(config.UpdateInterval); err != nil {
		glog.Fatalf("Invalid UpdateInterval %s (%s)", config.UpdateInterval, err)
		return
	}

	glog.Info("Starting daemon.")
	glog.V(1).Infof("Config: %v", config)

	m := newManager(&config)
	http.HandleFunc("/stats/", m.showStats)
	m.runWorkers()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		os.Interrupt,
		os.Kill,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

waitForSignals:
	for {
		sig := <-interrupt
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT:
			glog.Infof("Exit for interrupt: %v.", sig)
			m.stopWorkers()
			break waitForSignals
		case syscall.SIGINT, syscall.SIGTERM:
			glog.Infof("Exit immediately.")
			return
		}

	}

	glog.Infof("Waiting for all workers.")
	// TODO timeout
	m.waitForAllWorkers()
	glog.Infof("All workers finished.")
}

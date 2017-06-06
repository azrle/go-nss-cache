package main

import (
	"fmt"
	"github.com/azrle/go-nss-cache/client"
	"github.com/golang/glog"
	"net/http"
	"sync"
)

type manager struct {
	config  *Config
	workers []*worker
	wg      sync.WaitGroup
}

func newManager(config *Config) *manager {
	var accountTypes = [...]string{"users", "groups", "sshkeys"}
	var c client.Client

	switch config.Source {
	case "gce_metadata":
		c = client.NewGCEMetadatClient(config.GCEMetadata)
	default:
		glog.Fatalf("Unknown data source: %s\n", config.Source)
		return nil
	}

	m := &manager{
		workers: make([]*worker, 0),
		config:  config,
	}
	for _, name := range accountTypes {
		w := newWorker(name, config, c)
		m.workers = append(m.workers, w)
	}

	return m
}

func (m *manager) runWorkers() {
	for _, w := range m.workers {
		m.wg.Add(1)
		go func(w *worker) {
			glog.Infof("Worker (%s) is going to start.", w.name)
			w.run()
			m.wg.Done()
			glog.Infof("Worker (%s) is done.", w.name)
		}(w)
	}
}

func (m *manager) stopWorkers() {
	for wid := range m.workers {
		glog.V(1).Infof("Send stop signal to worker (%s).", m.workers[wid].name)
		go m.workers[wid].stop()
	}
}

func (m *manager) waitForAllWorkers() {
	m.wg.Wait()
}

func (m *manager) showStats(w http.ResponseWriter, r *http.Request) {
	for _, worker := range m.workers {
		fmt.Fprintf(w, "Worker %s status: %s\n", worker.name, worker.getStatus())
	}
}

package main

import (
	"fmt"
	"github.com/azrle/go-nss-cache/account"
	"github.com/azrle/go-nss-cache/client"
	"github.com/golang/glog"
	"sync"
	"time"
)

const (
	READY_STATUS    = "READY"
	RUNNING_STATUS  = "RUNNING"
	FETCH_STATUS    = "FETCH"
	SYNC_STATUS     = "SYNC"
	STOPPING_STATUS = "STOPPING"
	STOPPED_STATUS  = "STOPPED"
)

type worker struct {
	stopChan          chan struct{}
	name              string
	client            client.Client
	config            *Config
	status            string
	statusLock        sync.RWMutex
	heartbeat         time.Time
	lastError         error
	successiveFailure int
	lastSuccess       time.Time
}

func (w *worker) swapStatus(before, after string) error {
	glog.V(2).Infof("Worker (%s) is switching from %s to %s.", w.name, before, after)
	w.statusLock.Lock()
	defer w.statusLock.Unlock()
	if w.status == before {
		w.status = after
		return nil
	}
	return fmt.Errorf("Before getting into %s, the status should be %s, but it is %s.",
		after, before, w.status,
	)
}

func (w *worker) setStatus(status string) {
	glog.V(2).Infof("Worker (%s) is setting status to %s.", w.name, status)
	w.statusLock.Lock()
	defer w.statusLock.Unlock()
	w.status = status
}

func (w *worker) getStatus() string {
	w.statusLock.RLock()
	defer w.statusLock.RUnlock()
	return w.status
}

func (w *worker) fetchAccount() (account account.Account, err error) {
	glog.V(2).Infof("Worker (%s) starts fetching...", w.name)
	switch w.name {
	case "users":
		account, err = w.client.GetUsers()
	case "groups":
		account, err = w.client.GetGroups()
	case "sshkeys":
		account, err = w.client.GetSSHKeys()
	default:
		return nil, fmt.Errorf("Unknown worker name %s", w.name)
	}
	glog.V(2).Infof("worker (%s) result: %v", w.name, account)
	return
}

func (w *worker) updateStats(err error) {
	w.heartbeat = time.Now()
	if err != nil {
		glog.Errorf("Worker (%s) got an error: %s", w.name, err)
		w.lastError = err
		w.successiveFailure += 1
		return
	}
	w.lastSuccess = time.Now()
	w.successiveFailure = 0
	w.lastError = nil
}

func (w *worker) doMain() (keepGoing bool) {
	w.updateStats(func() (err error) {
		var acc account.Account
		if err = w.swapStatus(RUNNING_STATUS, FETCH_STATUS); err != nil {
			return
		}
		if acc, err = w.fetchAccount(); err != nil {
			return
		}
		if err = w.swapStatus(FETCH_STATUS, SYNC_STATUS); err != nil {
			return
		}
		if err = acc.UpdateNSSCache(); err != nil {
			return
		}
		if err = w.swapStatus(SYNC_STATUS, RUNNING_STATUS); err != nil {
			return
		}
		return
	}())
	if status := w.getStatus(); status != STOPPED_STATUS || status != STOPPING_STATUS {
		w.setStatus(RUNNING_STATUS)
		return true
	}
	return false
}

func (w *worker) run() {
	glog.V(1).Infof("Worker (%s) is going to run.", w.name)
	defer w.setStatus(STOPPED_STATUS)

	w.setStatus(RUNNING_STATUS)
	for w.doMain() {
		glog.V(2).Infof("Worker (%s) is going to sleep %v.", w.name, w.config.UpdateInterval)
		select {
		case <-w.stopChan:
			return
		case <-time.After(w.config.UpdateIntervalDuration):
			glog.V(2).Infof("Worker (%s) continues to work.", w.name)
			// contine to work
		}
	}
}

func (w *worker) stop() {
	glog.V(1).Infof("Worker (%s) is asked to stop.", w.name)
	w.setStatus(STOPPING_STATUS)
	select {
	case w.stopChan <- struct{}{}:
	default:
		// buffered channel
		// non-blocking even if called multiple times
	}
}

func newWorker(name string, config *Config, client client.Client) *worker {
	return &worker{
		name:     name,
		stopChan: make(chan struct{}, 1),
		client:   client,
		config:   config,
		status:   READY_STATUS,
	}
}

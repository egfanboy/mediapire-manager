package connectivity

import (
	"context"
	"sync"

	"github.com/egfanboy/mediapire-manager/internal/consul"
	"github.com/egfanboy/mediapire-manager/internal/media"
	"github.com/rs/zerolog/log"
)

type nodeConnectivityWatcher struct {
	mu    sync.RWMutex
	nodes map[string]bool
}

func (w *nodeConnectivityWatcher) WatchNode(nodeName, nodeId string) {
	serviceRemoved := make(chan struct{})

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.nodes[nodeId]; ok {
		log.Debug().Msgf("Node %s already being watched.", nodeId)
	} else {
		log.Debug().Msgf("Registering watcher for node %s.", nodeId)
		w.nodes[nodeId] = true
		go consul.WatchService(nodeName, serviceRemoved)
		go w.handleServiceRemoval(nodeId, serviceRemoved)
	}

}

func (w *nodeConnectivityWatcher) handleServiceRemoval(nodeId string, serviceRemoved <-chan struct{}) {
	<-serviceRemoved

	log.Info().Msgf("Node %s was removed", nodeId)
	ctx := context.Background()
	syncService, err := media.NewMediaSyncService(ctx)
	if err != nil {
		return
	}

	err = syncService.HandleRemovedNode(ctx, nodeId)
	if err != nil {
		log.Err(err).Msgf("Could not remove media from node %s", nodeId)
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	delete(w.nodes, nodeId)
}

var watcher = &nodeConnectivityWatcher{mu: sync.RWMutex{}, nodes: make(map[string]bool)}

func WatchNode(nodeName, nodeId string) {
	watcher.WatchNode(nodeName, nodeId)
}

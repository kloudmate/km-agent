package rpc

import (
	"fmt"
	"log"
	"time"

	"github.com/kloudmate/polylang-detector/detector"
	// dependency to polylang-detector for rpc calls
)

type RPCHandler struct{}

// PushDetectionResults receives a batch of ContainerInfo structs from a client
// and stores them in the in-memory cache.
func (h *RPCHandler) PushDetectionResults(results []detector.ContainerInfo, reply *string) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	log.Printf("Received a batch of %d detection results via RPC.", len(results))
	for _, info := range results {
		key := fmt.Sprintf("%s/%s/%s", info.PodName, info.Namespace, info.ContainerName)
		DetectionCache[key] = info
		fmt.Printf("Stored result for container '%s'.\n", info.ContainerName)
	}
	*reply = fmt.Sprintf("Successfully processed %d results and stored in cache.", len(results))
	return nil
}

// GetDetectionResults retrieves all stored results from the cache.
// This function is now for internal server use only.
func GetDetectionResults() []detector.ContainerInfo {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	var allResults []detector.ContainerInfo
	for _, info := range DetectionCache {
		allResults = append(allResults, info)
	}
	return allResults
}

// AutoCleanDetectionResults performs auto cleanup of laguage detection result every 5 minute
func AutoCleanDetectionResults() {
	t := time.NewTicker(time.Second * 90)
	defer t.Stop()
	for {
		<-t.C
		fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
		cacheMutex.Lock()
		newCache := make(map[string]detector.ContainerInfo)
		for key, container := range DetectionCache { // Use key from DetectionCache
			if !container.DetectedAt.Before(fiveMinutesAgo) {
				newCache[key] = container
			}
		}
		DetectionCache = newCache
		cacheMutex.Unlock()
		log.Printf("Cache cleanup: %d entries remaining", len(newCache))
	}
}

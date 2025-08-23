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
	// You can add logic here to clear the cache if needed
	// For example: detectionCache = make(map[string]ContainerInfo)
	return allResults
}

// AutoCleanDetectionResults performs auto cleanup of laguage detection result every 5 minute
func AutoCleanDetectionResults() {
	t := time.NewTicker(time.Second * 90)
	defer t.Stop()
	for {
		var usableContainers []detector.ContainerInfo
		// time threshold to remove any entry older than 5 minutes (5 minutes)
		fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
		for _, container := range DetectionCache {
			if !container.DetectedAt.Before(fiveMinutesAgo) {
				usableContainers = append(usableContainers, container)
			}
		}
		newCache := make(map[string]detector.ContainerInfo)
		for k, v := range usableContainers {
			newCache[string(rune(k))] = v
		}
		cacheMutex.Lock()
		DetectionCache = newCache
		cacheMutex.Unlock()

		<-t.C

	}
}

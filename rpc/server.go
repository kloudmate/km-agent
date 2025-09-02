package rpc

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"

	"github.com/kloudmate/polylang-detector/detector"
)

var (
	DetectionCache = make(map[string]detector.ContainerInfo)
	cacheMutex     sync.Mutex
)

// StartRpcServer starts the RPC server.
func StartRpcServer() {
	// Register the RPC handler
	rpc.Register(new(RPCHandler))

	// Listen for incoming connections on a specific port
	addr := ":" + os.Getenv("KM_CFG_UPDATER_RPC_ADDR")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error starting RPC server: %v", err)
	}
	defer listener.Close()

	go AutoCleanDetectionResults()
	// Accept connections and serve them concurrently
	log.Printf("RPC server listening on port %s\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

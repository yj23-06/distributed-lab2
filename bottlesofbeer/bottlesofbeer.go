package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)

// Global command-line flags
var (
	thisPort string
	nextAddr string
	startN   int
)

// Request/Response stub structure for RPC communication
type SongRequest struct {
	Bottles int
}

type SongResponse struct{}

// BuddyService handles incoming RPC calls from the previous buddy in the ring
type BuddyService struct{}

// SingVerse processes the current bottle count, prints the verse, and passes the turn to the next buddy
func (b *BuddyService) SingVerse(req SongRequest, resp *SongResponse) error {
	n := req.Bottles

	if n <= 0 {
		fmt.Println("No more bottles of beer on the wall, no more bottles of beer.")
		fmt.Println("Go to the store and buy some more, 99 bottles of beer on the wall!")
		os.Exit(0)
		return nil
	}

	// Print current verse
	if n == 1 {
		fmt.Println("1 bottle of beer on the wall, 1 bottle of beer.")
		fmt.Println("Take one down, pass it around, no more bottles of beer on the wall!")
	} else {
		fmt.Printf("%d bottles of beer on the wall, %d bottles of beer.\n", n, n)
		fmt.Printf("Take one down, pass it around, %d bottles of beer on the wall!\n\n", n-1)
	}

	time.Sleep(500 * time.Millisecond) // Pause for readability

	// Pass the turn to the next buddy asynchronously
	go passTurnToNext(n - 1)
	return nil
}

// passTurnToNext dials the next buddy in the ring and invokes SingVerse using client.Go
func passTurnToNext(remainingBottles int) {
	// Retry connection until the next buddy is active
	var client *rpc.Client
	var err error

	for i := 0; i < 10; i++ {
		client, err = rpc.Dial("tcp", nextAddr)
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if err != nil {
		fmt.Printf("Error connecting to next buddy at %s: %v\n", nextAddr, err)
		return
	}
	defer client.Close()

	req := SongRequest{Bottles: remainingBottles}
	var resp SongResponse

	// Call next buddy asynchronously using client.Go to avoid blocking thread chains
	call := client.Go("BuddyService.SingVerse", req, &resp, nil)
	<-call.Done // Wait for transmission completion
}

func main() {
	flag.StringVar(&thisPort, "this", "8030", "Port for this process to listen on")
	flag.StringVar(&nextAddr, "next", "127.0.0.1:8040", "IP:Port string for next member of the round")
	flag.IntVar(&startN, "n", 0, "Initial bottles of beer (launches song if > 0)")
	flag.Parse()

	// Register RPC service
	buddy := new(BuddyService)
	err := rpc.Register(buddy)
	if err != nil {
		fmt.Printf("RPC Register error: %v\n", err)
		return
	}

	// Listen on current port across all network interfaces
	listener, err := net.Listen("tcp", ":"+thisPort)
	if err != nil {
		fmt.Printf("Listener error on port %s: %v\n", thisPort, err)
		return
	}
	defer listener.Close()

	fmt.Printf("[Buddy listening on :%s] -> Next Buddy: %s\n", thisPort, nextAddr)

	// Accept incoming connections concurrently in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	// If initialized with -n > 0, kick off the singing chain
	if startN > 0 {
		fmt.Printf("\n--- Starting '99 Bottles of Beer' ring execution with %d bottles ---\n\n", startN)
		time.Sleep(1 * time.Second) // Small delay to allow all ring members to finish binding
		go passTurnToNext(startN)
	}

	// Keep main process active to process incoming requests
	select {}
}

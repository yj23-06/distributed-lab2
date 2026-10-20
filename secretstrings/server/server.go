package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/distributed2/secretstrings/stubs"
)

type SecretStringOperations struct{}

/** Super-Secret `reversing a string` method **/
func ReverseString(s string, i int) string {
	if i > 0 {
		time.Sleep(time.Duration(rand.Intn(i)) * time.Second)
	}
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Stage 1: Standard Reverse (includes random delay simulation)
func (s *SecretStringOperations) Reverse(req stubs.Request, res *stubs.Response) error {
	res.Message = ReverseString(req.Message, 2)
	return nil
}

// Stage 2: FastReverse for Premium Tier (instant execution, 0 delay)
func (s *SecretStringOperations) FastReverse(req stubs.Request, res *stubs.Response) error {
	res.Message = ReverseString(req.Message, 0)
	return nil
}

func main() {
	port := flag.String("port", "8030", "Port to listen on")
	flag.Parse()

	ops := new(SecretStringOperations)
	err := rpc.Register(ops)
	if err != nil {
		fmt.Printf("RPC Register error: %v\n", err)
		return
	}

	// Listen on all network interfaces (0.0.0.0:port) so external EC2 connections work
	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		fmt.Printf("Listener error: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("RPC Server listening on port :%s...\n", *port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}
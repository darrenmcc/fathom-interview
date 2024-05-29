package main

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestNewRTMPServer(t *testing.T) {
	server, err := NewRTMPServer(":0")
	if err != nil {
		t.Fatalf("failed to create RTMP server: %v", err)
	}
	if server.listener == nil {
		t.Fatal("expected a valid listener, got nil")
	}
	server.listener.Close()
}

func TestServerAcceptConnections(t *testing.T) {
	server, err := NewRTMPServer(":0")
	if err != nil {
		t.Fatalf("failed to create RTMP server: %v", err)
	}
	defer server.Shutdown()

	go server.Start()

	conn, err := net.Dial("tcp", server.listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect to RTMP server: %v", err)
	}
	conn.Close()

	time.Sleep(1 * time.Second)

	if server.shuttingDown {
		t.Fatal("server should not be shutting down")
	}
}

func TestServerIdleTimeoutShutdown(t *testing.T) {
	idleTimeout = 1 * time.Second
	server, err := NewRTMPServer(":0")
	if err != nil {
		t.Fatalf("failed to create RTMP server: %v", err)
	}
	defer server.Shutdown()

	go server.Start()

	time.Sleep(2 * time.Second)

	if !server.shuttingDown {
		t.Fatal("server should be shutting down due to idle timeout")
	}
}

func TestServerMaxLifetimeShutdown(t *testing.T) {
	maxLifetime = 1 * time.Second
	server, err := NewRTMPServer(":0")
	if err != nil {
		t.Fatalf("failed to create RTMP server: %v", err)
	}
	defer server.Shutdown()

	go server.Start()

	time.Sleep(2 * time.Second)

	if !server.shuttingDown {
		t.Fatal("server should be shutting down due to max lifetime")
	}
}

func TestServerGracefulShutdown(t *testing.T) {
	server, err := NewRTMPServer(":0")
	if err != nil {
		t.Fatalf("failed to create RTMP server: %v", err)
	}

	go server.Start()

	conn, err := net.Dial("tcp", server.listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect to RTMP server: %v", err)
	}
	defer conn.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		HandleStream(conn)
	}()

	server.mu.Lock()
	server.shuttingDown = true
	server.mu.Unlock()
	server.listener.Close()

	wg.Wait()
	server.Shutdown()

	if server.connections.HasActiveConnections() {
		t.Fatal("expected all connections to be closed")
	}
}

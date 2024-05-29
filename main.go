package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	// global vars instead of consts for testing
	maxLifetime = 6 * time.Hour
	idleTimeout = 15 * time.Minute
)

type connectionManager struct {
	count *int64
	wg    sync.WaitGroup
}

type RTMPServer struct {
	listener     net.Listener
	connections  *connectionManager
	mu           sync.Mutex
	shuttingDown bool
}

func NewRTMPServer(port string) (*RTMPServer, error) {
	log.Println("starting RTMP server on", port)

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	ln, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	var count int64
	return &RTMPServer{
		listener: ln,
		connections: &connectionManager{
			count: &count,
		},
	}, nil
}

func (s *RTMPServer) Start() {
	go func() {
		<-time.After(maxLifetime)
		log.Println("max lifetime reached, indicating readiness to shutdown")
		s.mu.Lock()
		s.shuttingDown = true
		s.mu.Unlock()
	}()

	go func() {
		ticker := time.NewTicker(idleTimeout)
		defer ticker.Stop()

		for range ticker.C {
			if s.ReadyToShutdown() {
				continue
			}
			if !s.connections.HasActiveConnections() {
				log.Println("no active connections for idle timeout, indicating readiness to shutdown")
				s.mu.Lock()
				s.shuttingDown = true
				s.mu.Unlock()
			}
		}
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.shuttingDown {
				break
			}
			log.Println("error accepting connection:", err)
			continue
		}
		s.mu.Lock()
		if s.shuttingDown {
			s.mu.Unlock()
			conn.Close()
			continue
		}
		s.connections.Add()
		s.mu.Unlock()
		go s.handleConnection(conn)
	}
}

func (s *RTMPServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shuttingDown {
		http.Error(w, "shutting down", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *RTMPServer) handleConnection(conn net.Conn) {
	defer s.connections.Done()
	HandleStream(conn)
	conn.Close()
}

func HandleStream(conn net.Conn) {
	// simulated handling of RTMP connection
	time.Sleep(5 * time.Second)
}

func (s *RTMPServer) Shutdown() {
	log.Println("shutting down RTMP server")
	s.mu.Lock()
	s.shuttingDown = true
	s.listener.Close()
	s.mu.Unlock()
	s.connections.Wait()
	log.Println("all connections closed, server shutdown complete")
}

func (s *RTMPServer) ReadyToShutdown() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shuttingDown
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server, err := NewRTMPServer(port)
	if err != nil {
		log.Fatal("failed to start RTMP server:", err)
	}

	go func() {
		http.HandleFunc("/health", server.healthCheckHandler)
		log.Println("HTTP server started on port")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	server.Start()

	// graceful shutdown on interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	server.Shutdown()
}

func (cm *connectionManager) HasActiveConnections() bool {
	return atomic.LoadInt64(cm.count) > 0
}

func (cm *connectionManager) Add() {
	_ = atomic.AddInt64(cm.count, 1)
	cm.wg.Add(1)
}
func (cm *connectionManager) Done() {
	_ = atomic.AddInt64(cm.count, -1)
	cm.wg.Done()
}
func (cm *connectionManager) Wait() {
	cm.wg.Wait()
}

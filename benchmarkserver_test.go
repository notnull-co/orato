package orato

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type benchmarkServer struct {
	expectedRequests int32
	requestMutex     sync.Mutex
	requestCount     int32
	finishedChannel  chan bool
	srv              *httptest.Server
}

func newBenchmarkServer(expectedRequests int32) *benchmarkServer {
	benchmarkServer := &benchmarkServer{
		expectedRequests: expectedRequests,
		finishedChannel:  make(chan bool, 1),
		requestMutex:     sync.Mutex{},
	}
	benchmarkServer.srv = httptest.NewServer(http.HandlerFunc(benchmarkServer.benchmark))
	return benchmarkServer
}

func (b *benchmarkServer) benchmark(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&b.requestCount, 1)
	if b.requestCount >= b.expectedRequests {
		b.finishedChannel <- true
		close(b.finishedChannel)
	}

	w.WriteHeader(http.StatusOK)
}

func (b *benchmarkServer) waitForServer() {
	backoff := 50 * time.Millisecond

	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", ":"+strings.Split(b.srv.URL, ":")[2], 1*time.Second)
		if err != nil {
			time.Sleep(backoff)
			continue
		}
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Fatalf("Server on URL %s not up after 10 attempts", b.srv.URL)
}

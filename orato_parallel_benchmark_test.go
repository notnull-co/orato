package orato

import (
	"testing"
)

func BenchmarkOratoParallelClient(b *testing.B) {
	requests := int32(1000)
	requestBuilders := 10

	payload := generateTestPayload(int(requests))
	testSrv := newBenchmarkServer(requests)
	testSrv.waitForServer()
	b.ResetTimer()

	oratoParallelClient(requestBuilders, payload, testSrv.srv.URL)
	defer testSrv.srv.Close()
	<-testSrv.finishedChannel
}

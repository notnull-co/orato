package orato

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPayload struct {
	ID   string
	Name string
}

func (t *testPayload) Identifier() string {
	return t.ID
}

func TestOratoParallelClient(t *testing.T) {
	requests := 105
	requestBuilders := 2

	benchmarkSrv := newBenchmarkServer(int32(requests))
	benchmarkSrv.waitForServer()
	defer benchmarkSrv.srv.Close()

	payload := generateTestPayload(requests)
	responseChannel := oratoParallelClient(requestBuilders, payload, benchmarkSrv.srv.URL)

	var responses []*OratoResponse
	for response := range responseChannel {
		assert.Nil(t, response.Error)
		assert.NotEmpty(t, response.ID)
		assert.NotEmpty(t, response.Response)
		responses = append(responses, response)
	}

	assert.Equal(t, requests, len(responses))
}

func TestOratoParallelClientWithError(t *testing.T) {
	requests := 105
	requestBuilders := 2

	payload := generateTestPayload(requests)
	responseChannel := oratoParallelClient(requestBuilders, payload, "\n")

	var responses []*OratoResponse
	for response := range responseChannel {
		assert.NotNil(t, response.Error)
		assert.NotEmpty(t, response.ID)
		assert.Nil(t, response.Response)
		responses = append(responses, response)
	}

	assert.Equal(t, requests, len(responses))
}

func oratoParallelClient(routines int, payload []*testPayload, url string) chan *OratoResponse {
	oratoClient := NewHttpClient(&http.Client{})

	parallel := NewParallelClient[*testPayload](oratoClient)

	return parallel.Do(func(p *testPayload) (*http.Request, error) {
		body, err := json.Marshal(p)

		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

		if err != nil {
			return nil, err
		}

		return req, nil
	}, routines, payload)
}

func generateTestPayload(count int) []*testPayload {
	payloads := []*testPayload{}
	for i := 0; i < count; i++ {
		payload := &testPayload{
			ID:   fmt.Sprintf("id-%d", i),
			Name: fmt.Sprintf("name-%d", i),
		}
		payloads = append(payloads, payload)
	}
	return payloads
}

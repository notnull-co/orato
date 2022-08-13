package orato

import "net/http"

type Identifiable interface {
	Identifier() string
}

type OratoParallel[T Identifiable] struct {
	client *OratoClient
}

func NewParallelClient[T Identifiable](client *OratoClient) *OratoParallel[T] {
	return &OratoParallel[T]{
		client: client,
	}
}

func (op *OratoParallel[T]) Do(fn func(T) (*http.Request, error), routines int, payload []T) chan *OratoResponse {
	chunkSize := (len(payload) + routines - 1) / routines

	payloadLen := len(payload)
	channels := []chan *OratoResponse{}
	response := make(chan *OratoResponse, len(payload))

	for i := 0; i < payloadLen; i += chunkSize {
		end := i + chunkSize

		if end > payloadLen {
			end = payloadLen
		}

		channels = append(channels, op.do(payload[i:end], op.doRequestFn(fn)))
	}

	go func() {
		defer close(response)
		for _, channel := range channels {
			for res := range channel {
				response <- res
			}
		}
	}()

	return response
}

func (op *OratoParallel[T]) do(payload []T, fn func(*OratoClient, T) chan *OratoResponse) chan *OratoResponse {
	response := make(chan *OratoResponse, len(payload))
	go func() {
		defer close(response)
		channels := make([]chan *OratoResponse, len(payload))
		for idx, req := range payload {
			channels[idx] = fn(op.client, req)
		}
		for _, channel := range channels {
			for res := range channel {
				response <- res
			}
		}
	}()
	return response
}

func (op *OratoParallel[T]) doRequestFn(fn func(T) (*http.Request, error)) func(c *OratoClient, p T) chan *OratoResponse {
	return func(c *OratoClient, p T) chan *OratoResponse {
		var response *OratoResponse
		chann := make(chan *OratoResponse, 1)
		defer func() {
			if response != nil {
				chann <- response
				close(chann)
			}
		}()

		req, err := fn(p)
		if err != nil {
			response = &OratoResponse{
				Error: err,
				ID:    p.Identifier(),
			}
			return chann
		}

		return c.Do(req, p.Identifier())
	}
}

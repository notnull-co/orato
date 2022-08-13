package orato

import (
	"net/http"
)

type OratoClient struct {
	*http.Client
}

type OratoResponse struct {
	Response *http.Response
	ID       string
	Error    error
}

func NewHttpClient(httpClient *http.Client) *OratoClient {
	return &OratoClient{
		Client: httpClient,
	}
}

func (oratoRequest *OratoClient) Do(r *http.Request, ID ...string) chan *OratoResponse {
	chann := make(chan *OratoResponse, 1)
	go func() {
		defer close(chann)
		response, err := oratoRequest.Client.Do(r)
		var id string
		if len(ID) > 0 {
			id = ID[0]
		}
		chann <- &OratoResponse{
			Response: response,
			Error:    err,
			ID:       id,
		}
	}()
	return chann
}

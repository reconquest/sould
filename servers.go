package main

import (
	"io/ioutil"
	"net/http"
)

// SecondaryServer is representation of slave sould server.
type SecondaryServer string

// ExecuteRequest creates and sends HTTP request basing on given propagatable
// request variable to given slave server using given http client.
func (slave SecondaryServer) ExecuteRequest(
	request PropagatableRequest,
	httpResource *http.Client,
) *ServerResponse {
	httpRequest, err := request.GetHTTPRequest(slave)
	if err != nil {
		return &ServerResponse{
			Slave:        slave,
			ErrorRequest: err,
		}
	}

	response, err := httpResource.Do(httpRequest)
	if err != nil {
		return &ServerResponse{
			Slave:        slave,
			ErrorRequest: err,
		}
	}

	body, err := ioutil.ReadAll(response.Body)

	return &ServerResponse{
		Slave:          slave,
		Status:         response.Status,
		StatusCode:     response.StatusCode,
		HeaderXError:   response.Header.Get("X-Error"),
		HeaderXSuccess: response.Header.Get("X-Success"),
		ResponseBody:   string(body),
		ErrorReceive:   err,
	}
}

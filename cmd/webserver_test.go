package cmd

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestStartWebServer starts the web server,
func TestStartWebServer(t *testing.T) {
	// Start web server
	stop := make(chan struct{})

	// start webserver in a go routine
	go func() {
		StartWebServer(stop)
	}()

	// wait for server to start
	time.Sleep(1 * time.Second)

	// send get request to /podhistory using http client
	resp, err := http.Get("http://localhost:8080/podhistory")
	assert.NoError(t, err, "failed to send get request to /podhistory")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "response status code is not 200")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "response content type is not application/json")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err, "failed to read response body")
	assert.Equal(t, "Pod History", string(body), "response body is not Pod History")

	// send stop signal
	stop <- struct{}{}
}

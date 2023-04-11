package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestStartWebServer starts the web server,
// currently only runs a dummy test to check if the server starts
func TestStartWebServer(t *testing.T) {

	withTestPodHistoryDB(t, func(phdb *PodHistoryDB, t *testing.T) {

		// Start web server
		stop := make(chan struct{})

		// start webserver in a go routine
		go func() {
			StartWebServer(phdb, stop)
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
		assert.JSONEq(t, `{"children":[],"name":"pods"}`, string(body), "response body is not as expected")

		// send get request to to / to get index.html
		resp, err = http.Get("http://localhost:8080/podhistory")
		assert.NoError(t, err, "failed to send get request to /")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "response status code is not 200")
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "response content type is not text/html; charset=utf-8")
		// print contents of the response body
		body, err = ioutil.ReadAll(resp.Body)
		assert.NoError(t, err, "failed to read response body")
		fmt.Println(string(body))

		// send stop signal
		stop <- struct{}{}

	})

}

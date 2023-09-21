package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
    uuid_generate "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type JSON map[string]interface{}

var jwtToken string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjM0NX0.0JbSlLakm5e2a4zkpTDsVGRCV_YMvyK7lWga4C6t8WQ"
var jwtTokenWithoutUserId string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJub3RfdXNlciI6MTIzNDV9.R8Ze0IKMUdY5N7oq5kYpkPTL_la5ZfLz-3wZolYbCqo"

func Test_main(t *testing.T) {
	port := 40000 + int(rand.Int31n(10000))
	os.Args = []string{"app", "--secret=123", "--listen=" + "localhost:"+strconv.Itoa(port), "--dsn=host=localhost port=5532 user=event password=9ju17UI6^Hvk dbname=micro_events sslmode=disable"}

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
	}()

	waitForHTTPServerStart(port)
	time.Sleep(time.Second)
    client := &http.Client{}

	{
	    url := fmt.Sprintf("http://localhost:%d/ping", port)
	    req, err := getRequest(url)
        resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "pong", string(body))
	}

	var uuid string

	userId := uuid_generate.New().String()

	{
        url :=  fmt.Sprintf("http://localhost:%d/api/v1/events", port)
        req, err := postRequest(url, `{"user_id": "`+fmt.Sprint(userId)+`","type": "test"}`)
        resp, err := client.Do(req)

		require.NoError(t, err)
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		var requestData JSON
		err = json.Unmarshal(respBody, &requestData)
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		assert.NotEmpty(t, requestData["uuid"])
		uuid = requestData["uuid"].(string)
	}

	{
        url :=  fmt.Sprintf("http://localhost:%d/api/v1/events/%s", port, uuid)
        req, err := postRequest(url, `{"status": "done"}`)
        resp, err := client.Do(req)

        require.NoError(t, err)
        defer resp.Body.Close()
        respBody, err := io.ReadAll(resp.Body)
        var requestData JSON
        err = json.Unmarshal(respBody, &requestData)
        assert.NoError(t, err)
        assert.Equal(t, uuid, requestData["uuid"])
        assert.Equal(t, 200, resp.StatusCode)
	}

	{
        url := fmt.Sprintf("http://localhost:%d/api/v1/events/users/%s", port, userId)
        req, err := getRequest(url)
        resp, err := client.Do(req)

        require.NoError(t, err)
        defer resp.Body.Close()
        respBody, err := io.ReadAll(resp.Body)
        var requestData JSON
        err = json.Unmarshal(respBody, &requestData)
        assert.NoError(t, err)
        assert.Equal(t, 200, resp.StatusCode)
    }

    {
        url :=  fmt.Sprintf("http://localhost:%d/api/v1/events/%s", port, uuid)
        req, err := postRequest(url, `{"fake": "done"}`)
        resp, err := client.Do(req)

        require.NoError(t, err)
        defer resp.Body.Close()
        respBody, err := io.ReadAll(resp.Body)
        var requestData JSON
        err = json.Unmarshal(respBody, &requestData)
        assert.NoError(t, err)
        assert.Equal(t, "error", requestData["status"])
        assert.Equal(t, 400, resp.StatusCode)
    }
}


func Test_Fail_Auth(t *testing.T) {
	port := 40000 + int(rand.Int31n(10000))
	os.Args = []string{"app", "--secret=123", "--listen=" + "localhost:"+strconv.Itoa(port), "--dsn=host=localhost port=5532 user=event password=9ju17UI6^Hvk dbname=micro_events sslmode=disable"}

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
	}()

	waitForHTTPServerStart(port)
	time.Sleep(time.Second)
    client := &http.Client{}

	{
	    url := fmt.Sprintf("http://localhost:%d/api/v1/events/users/1", port)
	    req, _ := getRequest(url)

	    req, err := http.NewRequest("GET", url, nil)
        req.Header.Add("Api-Token", "InvalidToken")

        resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 401, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid token\n", string(body))
	}

	{
        url := fmt.Sprintf("http://localhost:%d/api/v1/events/users/1", port)
        req, _ := getRequest(url)

        req, err := http.NewRequest("GET", url, nil)
        req.Header.Add("Api-Token", jwtTokenWithoutUserId)

        resp, err := client.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()
        assert.Equal(t, 401, resp.StatusCode)
        body, err := io.ReadAll(resp.Body)
        assert.NoError(t, err)
        assert.Equal(t, "user_id not found\n", string(body))
    }
}

// func Test_Fail_Run_UnknownFlag(t *testing.T) {
// 	os.Args = []string{"app", "--secret=123", "--unknown=1111111"}
//
// 	done := make(chan struct{})
// 	go func() {
// 		<-done
// 		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
// 		require.NoError(t, e)
// 	}()
//
// 	finished := make(chan struct{})
// 	go func() {
// 		main()
// 		close(finished)
// 	}()
//
// 	// defer cleanup because require check below can fail
// 	defer func() {
// 		close(done)
// 		<-finished
// 	}()
// }

func getRequest(url string) (*http.Request, error) {
    req, err := http.NewRequest("GET", url, nil)
    req.Header.Add("Api-Token", jwtToken)
    return req, err
}

func postRequest(url string, data string) (*http.Request, error) {
    req, err := http.NewRequest("POST", url, strings.NewReader(data))
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Api-Token", jwtToken)
    return req, err
}

func waitForHTTPServerStart(port int) {
	// wait for up to 10 seconds for server to start before returning it
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		if resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}

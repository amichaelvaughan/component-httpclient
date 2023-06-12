package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asecurityteam/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPDefaultComponent(t *testing.T) {
	cmp := &DefaultComponent{}
	conf := cmp.Settings()
	tr, err := cmp.New(context.Background(), conf)
	require.Nil(t, err)
	require.NotNil(t, tr)
}

func TestHTTPSmartComponentBadConfig(t *testing.T) {
	cmp := &SmartComponent{}
	conf := cmp.Settings()
	_, err := cmp.New(context.Background(), conf)
	require.NotNil(t, err)
}

var transportdConfig = `openapi: 3.0.0
x-transportd:
  backends:
    - app
  app:
    host: "http://app:8081"
    pool:
      ttl: "24h"
      count: 1
info:
  version: 1.0.0
  title: "Example"
  description: "An example"
  contact:
    name: Security Development
    email: secdev-external@atlassian.com
  license:
    name: Apache 2.0
    url: 'https://www.apache.org/licenses/LICENSE-2.0.html'
paths:
  /healthcheck:
    get:
      description: "Liveness check."
      responses:
        "200":
          description: "Success."
      x-transportd:
        backend: app
`

func TestHTTPSmartComponent(t *testing.T) {
	cmp := &SmartComponent{}
	conf := cmp.Settings()
	conf.OpenAPI = transportdConfig
	tr, err := cmp.New(context.Background(), conf)
	require.Nil(t, err)
	require.NotNil(t, tr)
}

func TestHTTP(t *testing.T) {
	src := settings.NewMapSource(map[string]interface{}{
		"httpclient": map[string]interface{}{
			"type": "DEFAULT",
		},
	})
	tr, err := New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, tr)

	src = settings.NewMapSource(map[string]interface{}{
		"httpclient": map[string]interface{}{
			"type": "SMART",
			"smart": map[string]interface{}{
				"openapi": transportdConfig,
			},
		},
	})
	tr, err = New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, tr)
	require.NotEqual(t, tr, http.DefaultTransport)

	src = settings.NewMapSource(map[string]interface{}{
		"httpclient": map[string]interface{}{
			"type": "MISSING",
		},
	})
	_, err = New(context.Background(), src)
	require.NotNil(t, err)
}

/*
Given a default http client with a default Content-Type of application/json
When a request is sent without a Content-Type header defined
Then the request is sent with a Content-Type header set to application/json
*/
func TestDefaultContentType(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "application/json", req.Header.Get(HeaderContentType))
		_, err := rw.Write([]byte(`OK`))
		if err != nil {
			return
		}
	}))
	defer server.Close()
	cmp := &DefaultComponent{}
	conf := cmp.Settings()
	tr, _ := cmp.New(context.Background(), conf)
	client := server.Client()
	client.Transport = tr
	req, _ := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(`{"hello": "world"}`))
	resp, _ := client.Do(req)
	assert.Equal(t, 200, resp.StatusCode)
}

/*
Given a default http client with a default Content-Type of application/json
When a request is sent with a Content-Type header set to application/jsonlines
Then the request is sent with a Content-Type header set to application/jsonlines
*/
func TestDefaultContentTypeOverrideable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "application/jsonlines", req.Header.Get(HeaderContentType))
		_, err := rw.Write([]byte(`OK`))
		if err != nil {
			return
		}
	}))
	defer server.Close()
	cmp := &DefaultComponent{}
	conf := cmp.Settings()
	tr, _ := cmp.New(context.Background(), conf)
	client := server.Client()
	client.Transport = tr
	req, _ := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(`{"hello": "world"}`))
	req.Header.Set(HeaderContentType, "application/jsonlines")
	resp, _ := client.Do(req)
	assert.Equal(t, 200, resp.StatusCode)
}

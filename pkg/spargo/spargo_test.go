package spargo

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"testing"
)

// More on round-trip: http://hassansin.github.io/Unit-Testing-http-client-in-Go

// RoundTripFunc
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestSparqlHandler(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testString)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	c := SPARQLClient{}
	c.Client = client
	c.ClientInit("http://example.com", testQuery)
	response := c.SPARQLGo()

	results := response.Results.Bindings
	if len(results) != 2 {
		t.Errorf("Anticipated length of 2, received %d", len(results))
	}

	expectedRes := []string{
		"http://the-fr.org/id/file-format/25",
		"OS/2 Bitmap",
		"http://the-fr.org/id/file-format/28",
		"CALS Compressed Bitmap",
	}

	var receivedRes []string
	for _, res := range results {
		for _, item := range res {
			receivedRes = append(receivedRes, item.Value)
		}
	}

	// FIXME: (Unless this works) Order of slices is not predictable in Golang.
	sort.Strings(receivedRes)
	sort.Strings(expectedRes)
	if reflect.DeepEqual(receivedRes, expectedRes) != true {
		t.Error("Result arrays are not equal")
	}
}

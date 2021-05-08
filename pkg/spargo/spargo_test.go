package spargo

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"testing"
)

// More on how we're using round-trip here from the awesome tutorial:
//
//    * http://hassansin.github.io/Unit-Testing-http-client-in-Go
//
//
// Transport specifies the mechanism by which individual HTTP requests
// are made.
//
// We want to intercept a request which transport normally receives and
// then send our own custom response. To do that we need to replace
// transport. To replace transport we need to implement round trip to
// send our own custom response. The response can be something we
// anticipate, or want to be able to handle as an exception.
//
// In short, we mock the request/response loop and make sure the sending
// and receiving happens as we expect. What happens in-between is
// in scope of the standard library and the Internet, but outside of
// scope of a unit test.

// RoundTripFunc describes an interface type that we will then implement.
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip implements Golang's roundtrip interface. From the
// documentation: "RoundTripper is an interface representing the ability
// to execute a single HTTP transaction, obtaining the Response for a
// given Request."
func (fn RoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid
// making real calls out to the Internet. Transport will then do
// whatever we request of it.
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

// TestSparqlHandler tests the request/receive capabilities of the
// package and simply makes sure that between posting a request
// and then receiving it and formatting it that the outcome is what
// was expected by the caller, i.e. the data is returned and parsed
// correctly by the library so that it can be used.
func TestSparqlHandler(t *testing.T) {
	httpClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(testString)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	sparql := SPARQLClient{}
	sparql.Client = httpClient
	sparql.ClientInit("http://example.com", testQuery)
	response := sparql.SPARQLGo()

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

	sort.Strings(receivedRes)
	sort.Strings(expectedRes)
	if reflect.DeepEqual(receivedRes, expectedRes) != true {
		t.Error("Result arrays are not equal")
	}
}

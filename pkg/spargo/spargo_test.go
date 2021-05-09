package spargo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strings"
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

// TestClientInit makes sure that we consistently get some sensible
// values when the SPARQLClient init function is called.
func TestClientInit(t *testing.T) {
	sparql := SPARQLClient{}

	sparql.ClientInit("", "")

	if sparql.BaseURL != "" {
		t.Errorf("ClientInit: BaseURL should be nil unless otherwise set, not %s", sparql.BaseURL)
	}

	if sparql.Query != "" {
		t.Errorf("ClientInit: Query should be nil unless otherwise set, not %s", sparql.Query)
	}

	if sparql.Agent != DefaultAgent {
		t.Errorf("ClientInit: Agent should be %s on first init", sparql.Agent)
	}

	if sparql.Accept != DefaultAccept {
		t.Errorf("ClientInit: Accept should be %s on first init", sparql.Accept)
	}

}

// TestSetupClient checks to make sure a http.Client interface is
// correctly provided to the endpoint type when setup is called.
func TestSetupClient(t *testing.T) {
	sparql := SPARQLClient{}
	if sparql.Client != nil {
		t.Error("SPARQLClient client should be nil before initialization")
	}
	setupClient(&sparql)
	emptyClientInterface := http.Client{}
	if reflect.TypeOf(sparql.Client) != reflect.TypeOf(&emptyClientInterface) {
		t.Error("SPARQLClient not setup with http.Client when setupClient() called")
	}
}

// spargoTests describes a row of data for testing with. The
// placeholders represent input and output values for our unit tests.
type spargoTests struct {
	statusCode        int
	okButFail         bool
	emptySPARQLresult bool
	responseValue     string
	resultsLen        int
	expectedRes       []string
}

const errorMessage = "Spargo: unexpected response from server: %d"

// spargoResults describes a table of inputs for our unit tests and
// their anticipated results values.
var spargoResults = []spargoTests{
	spargoTests{200, false, false, testString, 2, []string{"http://the-fr.org/id/file-format/25", "OS/2 Bitmap", "http://the-fr.org/id/file-format/28", "CALS Compressed Bitmap"}},
	spargoTests{200, false, true, testEmptyResult, 0, []string{}},
	spargoTests{300, false, false, "Unexpected test string", 0, nil},
	spargoTests{400, false, false, "Unexpected test string", 0, nil},
	spargoTests{418, false, false, "Unexpected test string", 0, nil},
	spargoTests{200, true, true, "{\"Parsing should fail gracefully", 0, []string{}},
	spargoTests{200, true, true, "Parsing should fail gracefully", 0, []string{}},
	spargoTests{200, true, true, "{\"No\": \"Real value\"}", 0, []string{}},
}

// TestSparqlHandler tests the request/receive capabilities of the
// package and simply makes sure that between posting a request
// and then receiving it and formatting it that the outcome is what
// was expected by the caller, i.e. the data is returned and parsed
// correctly by the library so that it can be used.
func TestSparqlHandler(t *testing.T) {
	for _, val := range spargoResults {

		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: val.statusCode,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(val.responseValue)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		})

		sparql := SPARQLClient{}
		sparql.Client = httpClient
		sparql.ClientInit("http://example.com", testQuery)
		response, err := sparql.SPARQLGo()

		results := response.Results.Bindings

		if val.emptySPARQLresult {
			if response.String() != (SPARQLResult{}.String()) {
				t.Errorf("Expected an empty SPARQL result, received; %s", response.String())
			}
		}

		if val.statusCode == 200 {

			if err != nil && !val.okButFail {
				t.Errorf("Expected 'nil' error from SPARQLGo, received: %s", err)
			}

			if val.okButFail {
				if !reflect.DeepEqual(response, SPARQLResult{}) {
					t.Errorf("Didn't receive an empty interface from SPARQLGo, received: %s", reflect.TypeOf(response))
				}
			}

			if len(results) != val.resultsLen {
				t.Errorf("Anticipated length of %d, received %d", val.resultsLen, len(results))
			}

			var receivedRes []string
			for _, res := range results {
				for _, item := range res {
					receivedRes = append(receivedRes, item.Value)
				}
			}

			if len(val.expectedRes) == 0 {
				// DeepEqual does not evaluate nil slices to be equal.
				if len(receivedRes) != 0 {
					t.Errorf("Expected results length 0 but got '%d': %s", len(receivedRes), receivedRes)
				}
				// Cannot use the tests to compare any further.
				return
			}

			sort.Strings(receivedRes)
			sort.Strings(val.expectedRes)
			if reflect.DeepEqual(receivedRes, val.expectedRes) != true {
				t.Errorf("Result arrays are not equal, received %s, expected %s", receivedRes, val.expectedRes)
			}
		}

		if val.statusCode != 200 {
			if err == nil {
				t.Errorf("Expected error from SPARQLGo, received: %s", err)
			}

			constructedError := fmt.Errorf(errorMessage, val.statusCode)

			if err.Error() != constructedError.Error() {
				t.Errorf("Expected specific error: '%s' from SPARQLGo, received: %s", constructedError, err)
			}

			if len(results) != 0 {
				t.Errorf("Results should not have been parsed by SPARGO")
			}
		}
	}
}

// RoundTripFuncError describes an interface type that we will then
// implement.
type RoundTripFuncError func(req *http.Request) *http.Response

// Mock error string.
const mockError = "Mock error..."

// RoundTripError implements Golang's roundtrip interface. In this
// version we want to simulate an error when trying to connect to a
// given SPARQL server.
func (fn RoundTripFuncError) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request), fmt.Errorf(mockError)
}

// NewTestClientError returns *http.Client with Transport replaced to
// avoid making real calls out to the Internet. Transport will then do
// whatever we request of it. With this version we mock an error in the
// call.
func NewTestClientError(fn RoundTripFuncError) *http.Client {
	return &http.Client{
		Transport: RoundTripFuncError(fn),
	}
}

// TestSparqlHandlerError tests the package when an error is returned
// by the http.Client.
func TestSparqlHandlerError(t *testing.T) {
	for _, val := range spargoResults {

		httpClient := NewTestClientError(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: val.statusCode,
				// Send response to be tested.
				Body: ioutil.NopCloser(bytes.NewBufferString(val.responseValue)),
				// Must be set to non-nil value or it panics.
				Header: make(http.Header),
			}
		})

		sparql := SPARQLClient{}
		sparql.Client = httpClient
		sparql.ClientInit("http://example.com", testQuery)
		response, err := sparql.SPARQLGo()

		if response.String() != (SPARQLResult{}.String()) {
			t.Errorf("Expected an empty SPARQL result, received; %s", response.String())
		}

		if err == nil {
			t.Errorf("Expected error from SPARQLGo, received: %s", err)
		}

		if !strings.Contains(err.Error(), mockError) {
			t.Errorf("Expected a mock error response from the call but it wasn't there: %s", err.Error())
		}
	}
}

// spargoTestsProof describes a row of data for testing with. The
// placeholders represent input and output values for our unit tests.
type spargoTestsProof struct {
	statusCode       int
	responseValue    string
	dataTypeSliceLen int
	langSliceLen     int
}

// spargoResultsProof describes a table of inputs for our unit tests and
// their anticipated results values.
var spargoResultsProof = []spargoTestsProof{
	spargoTestsProof{200, testString, 2, 2},
	spargoTestsProof{200, testEmptyResult, 0, 0},
}

// compare is a helper function for comparing two slices and failing the
// tests if they are not the same.
func compare(t *testing.T, arr1 []string, arr2 []string) {
	sort.Strings(arr1)
	sort.Strings(arr2)
	if !reflect.DeepEqual(arr1, arr2) {
		t.Errorf("Arrays do not match: '%s', and '%s'", arr1, arr2)
	}
}

// TestHumanEquivalence is a temporary function used to prove the
// equivalence between one of the interfaces that we are removing. For
// this test to work accurately we RoundTrip the result retrieved from
// "human" and compare the results with the response as it was originally
// received from the mock SPARQL endpoint.
func TestHumanEquivalence(t *testing.T) {
	for _, val := range spargoResultsProof {

		httpClient := NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: val.statusCode,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(val.responseValue)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		})

		sparql := SPARQLClient{}
		sparql.Client = httpClient
		sparql.ClientInit("http://example.com", testQuery)
		machineResponse, _ := sparql.SPARQLGo()

		if val.statusCode == 200 {

			var humanSparqlResponse SPARQLResult
			err := json.Unmarshal([]byte(machineResponse.Human), &humanSparqlResponse)
			if err != nil {
				t.Errorf("Expected 'nil' errors received; %s", err)
			}

			// Compare Values.
			var humanResValues []string
			for _, res := range humanSparqlResponse.Results.Bindings {
				for _, item := range res {
					humanResValues = append(humanResValues, item.Value)
				}
			}
			var machineResValues []string
			for _, res := range machineResponse.Results.Bindings {
				for _, item := range res {
					machineResValues = append(machineResValues, item.Value)
				}
			}
			compare(t, machineResValues, humanResValues)

			// Compare Language field, e.g. EN, DE.
			var humanResLang []string
			for _, res := range humanSparqlResponse.Results.Bindings {
				for _, item := range res {
					if item.Lang != "" {
						humanResLang = append(humanResLang, item.Lang)
					}
				}
			}
			var machineResLang []string
			for _, res := range machineResponse.Results.Bindings {
				for _, item := range res {
					if item.Lang != "" {
						machineResLang = append(machineResLang, item.Lang)
					}
				}
			}
			compare(t, machineResLang, humanResLang)
			if len(machineResLang) != val.langSliceLen ||
				len(humanResLang) != val.langSliceLen {
				t.Errorf("Lang slices are expected to be '%d' for this test, not '%d' and '%d'",
					val.langSliceLen,
					len(machineResLang),
					len(humanResLang),
				)
			}

			// Compare DataType, e.g. Decimal, DateTime.
			var humanResDataType []string
			for _, res := range humanSparqlResponse.Results.Bindings {
				for _, item := range res {
					if item.DataType != "" {
						humanResDataType = append(humanResDataType, item.DataType)
					}
				}
			}
			var machineResDataType []string
			for _, res := range machineResponse.Results.Bindings {
				for _, item := range res {
					if item.DataType != "" {
						machineResDataType = append(machineResDataType, item.DataType)
					}
				}
			}
			compare(t, machineResDataType, humanResDataType)
			if len(machineResDataType) != val.dataTypeSliceLen ||
				len(humanResDataType) != val.dataTypeSliceLen {
				t.Errorf("DataType slices are expected to be '%d' for this test, not '%d' and '%d'",
					val.dataTypeSliceLen,
					len(machineResDataType),
					len(humanResDataType),
				)
			}

			// Compare type, e.g. Literal, URI.
			var humanResType []string
			for _, res := range humanSparqlResponse.Results.Bindings {
				for _, item := range res {
					humanResType = append(humanResType, item.Type)
				}
			}
			var machineResType []string
			for _, res := range machineResponse.Results.Bindings {
				for _, item := range res {
					machineResType = append(machineResType, item.Type)
				}
			}
			compare(t, machineResType, humanResType)

		} else {
			t.Errorf("Response code different than 200: '%d'", val.statusCode)
		}

	}
}

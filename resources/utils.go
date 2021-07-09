package resources

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/clbanning/mxj/v2"
	"gitlab.com/sincap/sincap-common/json"
)

// Query2Map converts query parameters to a map
func Query2Map(values *url.Values) *map[string]interface{} {
	m := make(map[string]interface{}, len(*values))
	for key, value := range *values {
		m[key] = value[0]
	}
	return &m
}

// Body2Map converts request body to a map
func Body2Map(r *http.Request) (*map[string]interface{}, error) {
	m := make(map[string]interface{})
	// Decide json or xml
	var err error
	in, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	switch r.Header.Get("Content-Type") {
	case "application/json", "application/vnd.api+json":
		m, err = json.ToMap(&in)
	case "application/xml", "text/xml":
		m, err = mxj.NewMapXml(in) // unmarshal
		// remove root element to get object flat
		for _, v := range m {
			m = v.(map[string]interface{})
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return &m, err
}

// EndpointRequestTest requests to endpoint with given parameters.
func EndpointRequestTest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, []byte) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, nil
	}
	defer resp.Body.Close()

	return resp, respBody
}

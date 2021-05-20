package resources

import (
	"io/ioutil"
	"net/http"
	"net/url"

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

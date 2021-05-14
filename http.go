package eurekago

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type requestParam struct {
	URL                string
	Method             string
	Headers            map[string]string
	Body               string
	Username, Password string
}

func handleHttpRequest(reqParam *requestParam) (respBody []byte, status int, err error) {
	var (
		body io.Reader
		req  *http.Request
	)
	if reqParam.Body != "" {
		body = strings.NewReader(reqParam.Body)
	}
	fmt.Println(reqParam.Body)
	req, err = http.NewRequest(reqParam.Method, reqParam.URL, body)
	if err != nil {
		return
	}

	for k,v := range reqParam.Headers {
		req.Header.Set(k,v)
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if reqParam.Method == http.MethodGet {
		respBody, err = ioutil.ReadAll(resp.Body)
	}
	status = resp.StatusCode

	return
}

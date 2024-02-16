package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func Do(ctx context.Context, method, url string, request_header http.Header, response any) (err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return
	}

	for k, v := range request_header {
		req.Header.Set(k, v[0])
	}

	var resp *http.Response
	c := &http.Client{}
	resp, err = c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	if resp.Body != http.NoBody {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		if strings.Contains(resp.Header.Get("Content-Type"), "json") {
			err = json.Unmarshal(bodyBytes, response)
		}
	} else {
		err = errors.New("response body is empty")
	}

	return
}

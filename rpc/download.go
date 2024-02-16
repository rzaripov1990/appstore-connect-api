package rpc

import (
	"context"
	"io"
	"net/http"
	"os"
)

func FileSave(directory, name, url string, head http.Header) (err error) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	req.Header = head

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(directory+"/"+name, bodyBytes, 0666)
	if err != nil {
		return err
	}
	return
}

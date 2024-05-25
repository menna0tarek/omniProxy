package omnilink

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
)

func Request(url string, body []byte, key string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Println("Failed to send to the upstream : ", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("ApiKey %s", key))
	req.Header.Add("Content-Type", "application/json")
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)

}

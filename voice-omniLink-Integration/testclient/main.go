package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
)

func main() {
	// Handle incoming requests
	http.HandleFunc("/omnilink", handleRequest)

	// Start servers
	log.Fatal(http.ListenAndServe(":8088", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received")
	w.WriteHeader(200) // Send response
	//check request body
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	fmt.Println("Request body: ", body)
	// extract transactionId from response body
	// sample response body
	//  {"name":"testvoice","transactionId":"a8faba64-ebc1-423b-9d63-3c31aef313a7","to":{"SubscriberId":"01022876977","Phone":"01022876977"},"Overrides":{"voice":{"voiceCallbackUrl":"https://45d0-156-197-35-174.ngrok-free.app/callback"}},"Payload":{}}
	transactionID := regexp.MustCompile(`"transactionId":"(.*?)"`).FindStringSubmatch(body)[1]
	fmt.Println("Transaction ID: ", transactionID)

	w.Write([]byte{})
	time.Sleep(1 * time.Second)

	toWrite := `{"to":"201002071244","callId":"%s","statusCallbackEndpoint":"https://webhook-qc-omnilink.cequens.net:8082/webhooks/organizations/66056f29b197fe9d49f384b0/environments/66056f29b197fe9d49f384b6/voice/cequensVoice","redirectUrl":"http://localhost:8088/omnilink/callID=%s","ringingDuration":30,"executionPlan":"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n<begin>\n<play url=\"https://www.asterisksounds.org/sites/asterisksounds.org/files/sounds/en/extra/countdown.ogg\"></play>\n</begin>"}`

	// toWrite = `{"to":"201002071244","callId":"%s","statusCallbackEndpoint":"https://webhook-qc-omnilink.cequens.net:8082/webhooks/organizations/66056f29b197fe9d49f384b0/environments/66056f29b197fe9d49f384b6/voice/cequensVoice","redirectUrl":"http://localhost:8088/omnilink","ringingDuration":30,"executionPlan":"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n<begin>\n<play url=\"/var/lib/asterisk/sounds/asterisk-recording1\"></play>\n</begin>"}`

	toWrite = fmt.Sprintf(toWrite, transactionID, transactionID)
	doRequest("http://localhost:8089/callback", []byte(toWrite))
	fmt.Println("Create Callback request is done")
}

func doRequest(url string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Println("Failed to send to the upstream : ", err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(resp.StatusCode)

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("body to read", string(body))
	return nil
}

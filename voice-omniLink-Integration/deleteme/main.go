package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

func main() {
	http.HandleFunc("/requests", handler)
	log.Println("Server starting on port 9000...")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request", r.Method)
	log.Printf("Request headers: %v", r.Header)
	log.Printf("Request body: %v", r.Body)
	// Read the request body
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	log.Printf("Request body: %s", body)
	// extract from below body the transactionID
	// Request body: {"name":"testvoice","transactionId":"b7f9e4bd-39cc-4bf6-be9f-e5c27990a884","to":{"SubscriberId":"01022876977","Phone":"01022876977"},"Overrides":{"voice":{"voiceCallbackUrl":"https://45d0-156-197-35-174.ngrok-free.app/callback"}},"Payload":{}}
	transactionID := regexp.MustCompile(`"transactionId":"(.*?)"`).FindStringSubmatch(body)[1]
	log.Printf("Transaction ID: %s", transactionID)

	// Send 200 OK response
	w.WriteHeader(http.StatusOK)

	// Asynchronously call the upstream service
	go callUpstreamService(transactionID)
}

func callUpstreamService(transactionID string) {
	time.Sleep(time.Second)
	xmlContent := `<begin><play url='https://www.asterisksounds.org/sites/asterisksounds.org/files/sounds/en/extra/countdown.ogg'></play></begin>`
	xmlResponsefmt := `{"to":"201002071244","callId":"198d9723-8bcb-41b4-998b-8d04be2d068a","statusCallbackEndpoint":"https://localhost:9000//webhooks/organizations/%s/environments/66056f29b197fe9d49f384b6/voice/cequensVoice","redirectUrl":"https://localhost:9000/webhooks/organizations/66056f29b197fe9d49f384b0/environments/%s/jobs/66093ff5361949b8bb423e61/voice-call-exec/cequensVoice","ringingDuration":30,"executionPlan":"%s"}`
	xmlResponse := fmt.Sprintf(xmlResponsefmt, transactionID, transactionID, xmlContent)
	req, err := http.NewRequest("POST", "https://45d0-156-197-35-174.ngrok-free.app/callback", bytes.NewBufferString(xmlResponse))
	if err != nil {
		log.Printf("Error creating request: %s", err)
		return
	}

	req.Header.Set("Content-Type", "application/xml")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to upstream service: %s", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("callback service responded with status: %s", resp.Status)
}

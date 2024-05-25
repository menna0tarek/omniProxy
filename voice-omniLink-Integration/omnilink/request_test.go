package omnilink

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert" // Using a testing framework for better assertions
)

func TestOmnilinkRequestSuccess(t *testing.T) {
	// Construct the expected request body
	event := map[string]interface{}{
		"name":          "testvoice",
		"transactionId": "198d9723-8bcb-41b4-998b-8d04be2d068e",
		"to": map[string]string{
			"subscriberId": "201002071244",
			"phone":        "201002071244",
		},
		"overrides": map[string]interface{}{
			"voice": map[string]string{
				"voiceCallbackUrl": "http://95.177.166.217:8989/callback",
			},
		},
		"payload": map[string]interface{}{},
	}
	requestBody, err := json.Marshal(event)
	assert.NoError(t, err, "Failed to marshal event data")
	start_time := time.Now()
	out, err := Request("https://omnilink-api.apps.omnilink.k8sctl.qc01.cequens.local:8443/v1/events/trigger", requestBody, "a05a346016cd93fffe620a8f51e221a4")
	elapsed_time := time.Since(start_time)
	assert.NoError(t, err, "Failed to do the request ")
	fmt.Println("The request took : ", elapsed_time)
	fmt.Println(out)
}

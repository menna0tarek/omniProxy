package omnilink

import (
	"encoding/xml"
	"testing"
)

var s1 = `{"to":"201002071244","callId":"198d9723-8bcb-41b4-998b-8d04be2d068a","statusCallbackEndpoint":"https://webhook-qc-omnilink.cequens.net:8082/webhooks/organizations/66056f29b197fe9d49f384b0/environments/66056f29b197fe9d49f384b6/voice/cequensVoice","redirectUrl":"https://webhook-qc-omnilink.cequens.net:8082/webhooks/organizations/66056f29b197fe9d49f384b0/environments/66056f29b197fe9d49f384b6/jobs/66093ff5361949b8bb423e61/voice-call-exec/cequensVoice","ringingDuration":30,"executionPlan":"<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>\n<begin>\n  <say reps=\"1\" voice=\"Danielle\" speed=\"1\">Hello</say>\n</begin>"}`

var s2 = `{"to":"201002071244","callId":"198d9723-8bcb-41b4-998b-8d04be2d068a","statusCallbackEndpoint":"https://webhook-qc-omnilink.cequens.net:8082/webhooks/organizations/66056f29b197fe9d49f384b0/environments/66056f29b197fe9d49f384b6/voice/cequensVoice","redirectUrl":"https://webhook-qc-omnilink.cequens.net:8082/webhooks/organizations/66056f29b197fe9d49f384b0/environments/66056f29b197fe9d49f384b6/jobs/66093ff5361949b8bb423e61/voice-call-exec/cequensVoice","ringingDuration":30,"executionPlan":"<begin>\n  <say reps=\"1\" voice=\"Danielle\" speed=\"1\">Hello</say>\n</begin>"}`

func TestExtract(t *testing.T) {
	output := Extract(s2)
	// Optional: If you want to ensure the modified XML is valid, you can unmarshal and marshal it.
	var xmlData struct{}
	if err := xml.Unmarshal([]byte(output), &xmlData); err != nil {
		t.Errorf("Error unmarshaling XML: %v", err)
	}
	t.Log(xmlData)
	// marshaledXML, err := xml.MarshalIndent(xmlData, "", "  ")
	// if err != nil {
	// 	t.Errorf("Error marshaling XML: %v", err)
	// }
	// t.Log("Validated and indented XML:")
	// t.Log(string(marshaledXML))

}

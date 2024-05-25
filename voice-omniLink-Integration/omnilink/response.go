package omnilink

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Data struct {
	ExecutionPlan string `json:"executionPlan"`
	RedirectURL   string `json:"redirectUrl"`
}

func Extract(jsonData []byte) string {
	log.Println("[TRACE][OL.Extract] Extracting data from JSON")
	var data Data
	if err := json.Unmarshal(jsonData, &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return ""
	}

	// Prepare the XML modification
	insertion := fmt.Sprintf("  <redirect url=\"%s\"></redirect>", data.RedirectURL)
	modifiedXML := strings.Replace(data.ExecutionPlan, "</begin>", insertion+"\n</begin>", 1)

	// Output the modified XML
	log.Println("[DEBUG][OL.Extract] Modified XML:")
	fmt.Println(modifiedXML)
	return modifiedXML

}

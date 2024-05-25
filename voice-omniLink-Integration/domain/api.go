package domain

type CallbackRequestData struct {
	To                     string `json:"to"`
	CallID                 string `json:"callId"`
	StatusCallbackEndpoint string `json:"statusCallbackEndpoint"`
	RedirectURL            string `json:"redirectUrl"`
	RingingDuration        int    `json:"ringingDuration"`
	ExecutionPlan          string `json:"executionPlan"`
}

type RequestData struct {
	CallId string `json:"transactionId"`
	Name   string `json:"name"`
	To     struct {
		SubscriberId string `json:"subscriberId"`
		Phone        string `json:"phone"`
	} `json:"to"`
}

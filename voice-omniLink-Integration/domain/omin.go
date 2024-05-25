package domain

type OmnilinkRequestData struct {
	Name          string `json:"name"`
	TransactionId string `json:"transactionId"`
	To            ToData `json:"to"`
	Overrides     map[string]map[string]string
	Payload       map[string]interface{}
}

type ToData struct {
	SubscriberId string
	Phone        string
}

package domain

type DataPoint struct {
	Code  string `json:"code"`
	Value any    `json:"value"`
}

type Channel struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type Device struct {
	ID              string      `json:"id"`
	Category        string      `json:"category"`
	Name            string      `json:"name"`
	Status          []DataPoint `json:"status"`
	CodeNameMapping []Channel   `json:"code_name_mapping"`
}

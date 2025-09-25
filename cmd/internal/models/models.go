package models

type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Pass  string `json:"pass"`
}

type LoginPassData struct {
	Login    string `json:"login"`
	Pass     string `json:"pass"`
	Resource string `json:"resource"`
}

type Memo struct {
	TextData string `json:"text_data"`
}

type BinaryData struct {
	BinaryData []byte `json:"binary_data"`
}

type BankData struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
}

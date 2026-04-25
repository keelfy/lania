package requests

type EasyDonateCallbackRequest struct {
	PaymentID int64  `json:"payment_id"`
	Cost      int64  `json:"cost"`
	Customer  string `json:"customer"`
	Signature string `json:"signature"`
}

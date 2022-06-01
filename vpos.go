package vposgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lithammer/shortuuid"
)

const (
	sandboxURL         = "https://sandbox.vpos.ao/api/v1"
	productionURL      = "https://api.vpos.ao/api/v1"
	paymentTransaction = "payment"
	refundTransaction  = "refund"
)

type (
	VPOS struct {
		Token              string //Your VPOS Token
		PosID              int64  //Your GPO POS ID
		PaymentCallbackURL string //Your Payment Callback URL
		RefundCallbackURL  string //Your Refund Callback URL
		SupervisorCard     string //Your GPO Supervisor Card
		Environment        string //Your Environment, must be 'production' or 'sandbox'
	}

	PaymentTransaction struct {
		Type        string `json:"type"`
		PosID       int64  `json:"pos_id"`
		Mobile      string `json:"mobile"`
		Amount      string `json:"amount"`
		CallbackURL string `json:"callback_url"`
	}

	TransactionStatus struct {
		SecondsRemaining json.Number `json:"eta"`
		CreatedAt        string      `json:"inserted_at"`
	}
)

func NewVPOS(token string, posID int64, paymentCallbackURL, refundCallbackURL, supervisorCard, environment string) (*VPOS, error) {
	if !(environment == "production" || environment == "sandbox") {
		return &VPOS{}, errors.New("invalid environment")
	}

	return &VPOS{
		Token:              token,
		PosID:              posID,
		PaymentCallbackURL: paymentCallbackURL,
		RefundCallbackURL:  refundCallbackURL,
		SupervisorCard:     supervisorCard,
		Environment:        environment,
	}, nil
}

func GetStatusReason(code int64) (reason string, err error) {
	reason, ok := statusReason[code]
	if !ok {
		return "", errors.New("invalid status reason")
	}
	return
}

func (v *VPOS) TransactionRemainingTime(transactionID string) (result int64, err error) {
	url := fmt.Sprintf("%s/requests/%s", sandboxURL, transactionID)
	if v.Environment == "production" {
		url = fmt.Sprintf("%s/requests/%s", productionURL, transactionID)
	}

	response, err := httpGet(
		url,
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", v.Token),
		},
	)
	if err != nil {
		return
	}

	var status TransactionStatus
	if err = json.Unmarshal(response, &status); err != nil {
		return
	}

	floatResult, err := status.SecondsRemaining.Float64()
	if err != nil {
		return
	}

	return int64(floatResult), nil
}

func (v *VPOS) InitPaymentTransaction(transactionType, mobile, amount string) (transactionID, idempotencyKey, nonce string, timeRemaining int64, err error) {
	var callbackURL string

	if transactionType == paymentTransaction {
		callbackURL = v.PaymentCallbackURL
	} else if transactionType == refundTransaction {
		callbackURL = v.RefundCallbackURL
	} else {
		err = errors.New("invalid transaction type")
		return
	}

	url := fmt.Sprintf("%s/transactions", sandboxURL)
	if v.Environment == "production" {
		url = fmt.Sprintf("%s/transactions", productionURL)
	}

	idempotencyKey, nonce = shortUUID(), shortUUID()

	request := PaymentTransaction{
		Type:        transactionType,
		PosID:       v.PosID,
		Mobile:      mobile,
		Amount:      amount,
		CallbackURL: fmt.Sprintf("%s?nonce=%s", callbackURL, nonce),
	}

	_, headers, err := httpPost(
		url,
		map[string]string{
			"Authorization":   fmt.Sprintf("Bearer %s", v.Token),
			"Idempotency-Key": idempotencyKey,
		},
		request,
	)
	if err != nil {
		return
	}

	location := headers.Get("Location")
	if location == "" {
		err = errors.New("could not retrieve transaction ID from VPOS response")
		return
	}

	locationParts := strings.Split(location, "/")
	locationPartsSize := len(locationParts)
	if locationPartsSize == 0 {
		err = errors.New("could not retrieve transaction ID from VPOS response")
		return
	}

	transactionID = locationParts[locationPartsSize-1]
	timeRemaining, _ = v.TransactionRemainingTime(transactionID)

	return
}

func shortUUID() string {
	return shortuuid.New()
}

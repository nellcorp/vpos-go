package vposgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lithammer/shortuuid"
)

const (
	sandboxURL               = "https://sandbox.vpos.ao/api/v1"
	productionURL            = "https://api.vpos.ao/api/v1"
	paymentTransaction       = "payment"
	refundTransaction        = "refund"
	authorizationTransaction = "authorization"
	cancelationTransaction   = "cancelation"
)

type (
	VPOS struct {
		Token              string //Your VPOS Token
		PosID              int64  //Your GPO POS ID
		PaymentCallbackURL string //Your Payment Callback URL
		RefundCallbackURL  string //Your Refund Callback URL
		SupervisorCard     string //Your GPO Supervisor Card
		Environment        string //Your Environment, set "production" for production or "development" for sandbox environment
	}

	PaymentTransaction struct {
		Type        string `json:"type"`
		PosID       int64  `json:"pos_id"`
		Mobile      string `json:"mobile"`
		Amount      string `json:"amount"`
		CallbackURL string `json:"callback_url"`
	}

	RefundTransaction struct {
		Type                string `json:"type"`
		ParentTransactionID string `json:"parent_transaction_id"`
		CallbackURL         string `json:"call_back_url"`
	}

	TransactionStatus struct {
		SecondsRemaining json.Number `json:"eta"`
		CreatedAt        string      `json:"inserted_at"`
	}
)

func NewVPOS(posID int64, token, paymentCallbackURL, refundCallbackURL, supervisorCard, environment string) (*VPOS, error) {
	if !(environment == "production" || environment == "development") {
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

//transactionType is either 'payment' or 'authorization'
//if transactionType is 'payment', it will create a new payment transaction
//if transactionType is 'authorization', it will create a new authorization transaction
func (v *VPOS) PaymentTransaction(transactionType, mobile, amount string) (transactionID, idempotencyKey, nonce string, timeRemaining int64, err error) {
	if !(transactionType == paymentTransaction || transactionType == authorizationTransaction) {
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
		CallbackURL: fmt.Sprintf("%s?nonce=%s", v.PaymentCallbackURL, nonce),
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

//transactionType is either 'refund' or 'cancelation'
//if transactionType is 'refund', it will create a new refund transaction
//if transactionType is 'cancelation', it will create a new cancelation transaction for a previously accepted payment authorization
func (v *VPOS) RefundOrCancelation(transactionType, parent_transaction_id string) (transactionID, idempotencyKey, nonce string, timeRemaining int64, err error) {
	if !(transactionType == refundTransaction || transactionType == cancelationTransaction) {
		err = errors.New("invalid transaction type")
		return
	}

	url := fmt.Sprintf("%s/transactions", sandboxURL)
	if v.Environment == "production" {
		url = fmt.Sprintf("%s/transactions", productionURL)
	}

	idempotencyKey, nonce = shortUUID(), shortUUID()

	request := RefundTransaction{
		Type:                transactionType,
		ParentTransactionID: parent_transaction_id,
		CallbackURL:         fmt.Sprintf("%s?nonce=%s", v.RefundCallbackURL, nonce),
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

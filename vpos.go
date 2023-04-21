package vposgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lithammer/shortuuid"
)

const (
	baseURL                           = "https://vpos.ao/api/v1"
	TypePayment       TransactionType = "payment"
	TypeRefund        TransactionType = "refund"
	TypeAuthorization TransactionType = "authorization"
	TypeCancelation   TransactionType = "cancelation"
)

type (
	TransactionType string
	VPOS            struct {
		Token              string //Your VPOS Token
		PosID              int64  //Your GPO POS ID
		PaymentCallbackURL string //Your Payment Callback URL
		RefundCallbackURL  string //Your Refund Callback URL
		SupervisorCard     string //Your GPO Supervisor Card
	}

	PaymentTransaction struct {
		Type        TransactionType `json:"type"`
		PosID       int64           `json:"pos_id"`
		Mobile      string          `json:"mobile"`
		Amount      string          `json:"amount"`
		CallbackURL string          `json:"callback_url"`
	}

	AuthorizedPaymentTransaction struct {
		Type                TransactionType `json:"type"`
		ParentTransactionID string          `json:"parent_transaction_id"`
		Amount              string          `json:"amount"`
		CallbackURL         string          `json:"callback_url"`
	}

	RefundTransaction struct {
		Type                TransactionType `json:"type"`
		ParentTransactionID string          `json:"parent_transaction_id"`
		CallbackURL         string          `json:"call_back_url"`
	}

	TransactionStatus struct {
		SecondsRemaining json.Number `json:"eta"`
		CreatedAt        string      `json:"inserted_at"`
	}

	Transaction struct {
		ID                  string          `json:"id"`
		Amount              string          `json:"amount"`
		ClearingPeriod      string          `json:"clearing_period"`
		Mobile              string          `json:"mobile"`
		ParentTransactionID string          `json:"parent_transaction_id"`
		PosID               int64           `json:"pos_id"`
		Status              string          `json:"status"`
		StatusDatetime      string          `json:"status_datetime"`
		StatusReason        string          `json:"status_reason"`
		Type                TransactionType `json:"type"`
	}
)

func NewVPOS(posID int64, token, paymentCallbackURL, refundCallbackURL, supervisorCard string) *VPOS {
	return &VPOS{
		Token:              token,
		PosID:              posID,
		PaymentCallbackURL: paymentCallbackURL,
		RefundCallbackURL:  refundCallbackURL,
		SupervisorCard:     supervisorCard,
	}
}

func GetStatusReason(code int64) (reason string, err error) {
	reason, ok := statusReason[code]
	if !ok {
		return "", errors.New("invalid status reason")
	}
	return
}

func (v *VPOS) GetRequest(transactionID string) (result int64, err error) {
	url := fmt.Sprintf("%s/requests/%s", baseURL, transactionID)

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

// transactionType is either 'payment' or 'authorization'
// if transactionType is 'payment', it will create a new payment transaction
// if transactionType is 'authorization', it will create a new authorization transaction
func (v *VPOS) PaymentTransaction(transactionType TransactionType, mobile, amount string) (transactionID, idempotencyKey, nonce string, timeRemaining int64, err error) {
	if !(transactionType == TypePayment || transactionType == TypeAuthorization) {
		err = errors.New("invalid transaction type")
		return
	}

	url := fmt.Sprintf("%s/transactions/%s", baseURL, transactionID)

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
	timeRemaining, _ = v.GetRequest(transactionID)

	return
}

// It creates a new payment transaction from a previously accepted authorization transaction
func (v *VPOS) PaymentWithAuthorization(parent_transaction_id, amount string) (transactionID, idempotencyKey, nonce string, timeRemaining int64, err error) {
	url := fmt.Sprintf("%s/transactions/%s", baseURL, transactionID)

	idempotencyKey, nonce = shortUUID(), shortUUID()
	request := AuthorizedPaymentTransaction{
		Type:                TypePayment,
		ParentTransactionID: parent_transaction_id,
		Amount:              amount,
		CallbackURL:         fmt.Sprintf("%s?nonce=%s", v.PaymentCallbackURL, nonce),
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
	timeRemaining, _ = v.GetRequest(transactionID)

	return
}

// transactionType is either 'refund' or 'cancelation'
// if transactionType is 'refund', it will create a new refund transaction
// if transactionType is 'cancelation', it will create a new cancelation transaction for a previously accepted payment authorization
func (v *VPOS) RefundOrCancelation(transactionType TransactionType, parentTransactionID string) (transactionID, idempotencyKey, nonce string, timeRemaining int64, err error) {
	if !(transactionType == TypeRefund || transactionType == TypeCancelation) {
		err = errors.New("invalid transaction type")
		return
	}

	url := fmt.Sprintf("%s/transactions/%s", baseURL, transactionID)

	idempotencyKey, nonce = shortUUID(), shortUUID()

	request := RefundTransaction{
		Type:                transactionType,
		ParentTransactionID: parentTransactionID,
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
	timeRemaining, _ = v.GetRequest(transactionID)

	return
}

func (v *VPOS) GetTransaction(transactionID string) (transaction Transaction, err error) {
	url := fmt.Sprintf("%s/transactions/%s", baseURL, transactionID)

	response, err := httpGet(
		url,
		map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", v.Token),
		},
	)
	if err != nil {
		return
	}

	if err = json.Unmarshal(response, &transaction); err != nil {
		return
	}

	return
}

func shortUUID() string {
	return shortuuid.New()
}

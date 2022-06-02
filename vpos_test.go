package vposgo

import (
	"testing"
)

var (
	posID              = int64(0)
	token              = ""
	refundCallbackURL  = ""
	paymentCallbackURL = ""
	supervidorCard     = ""
	vpos               = NewVPOS(
		posID,
		token,
		paymentCallbackURL,
		refundCallbackURL,
		supervidorCard,
		"",
	)
)

func TestPaymentTransaction(t *testing.T) {
	_, _, _, _, err := vpos.PaymentTransaction("payment", "900111222", "123.45")
	if err != nil {
		t.Logf("something went wrong: %v", err)
		t.Fail()
	}
}

func TestFailPaymentTransactionWithInvalidToken(t *testing.T) {
	vpos.Token = "vpos-token"

	_, _, _, _, err := vpos.PaymentTransaction("payment", "900111222", "123.45")
	if err == nil {
		t.Log("payment should not be performed with that token, something went wrong!")
		t.Fail()
	}
}

func TestFailRefundTransaction(t *testing.T) {
	_, _, _, _, err := vpos.NewRefund("non-existent-transaction-id")
	if err != nil {
		t.Logf("something went wrong: %v", err)
		t.Fail()
	}
}

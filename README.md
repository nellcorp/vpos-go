# vPOS Go

This package facilitates the integration of your Golang project/application with the vPOS API Payment Gateway.

## Installation

```sh
go get github.com/nellcorp/vpos-go
```

## Configuration
This package requires you set up the following environment variables on your machine before
interacting with the API using this package:

| Variable | Description | Required |
| --- | --- | --- |
| `GPO_POS_ID` | The Point of Sale ID provided by EMIS | true |
| `GPO_SUPERVISOR_CARD` | The Supervisor card ID provided by EMIS | true |
| `MERCHANT_VPOS_TOKEN` | The API token provided by vPOS | true |
| `PAYMENT_CALLBACK_URL` | The URL that will handle payment notifications | false |
| `REFUND_CALLBACK_URL` | The URL that will handle refund notifications | false |

Given you have set up all the environment variables above with the correct information, you will now
be able to authenticate and communicate effectively with vPOS API Payment Gateway using this package. 

**Note:** During development, make sure to use a token created for the sandbox environment

The next section will show the various payment actions that can be performed by you, the merchant.

### How to instantiate vPOS
To create an instance of a vPOS merchant see argument table and a simple example below:

| Argument | Description | Type |
| --- | --- | --- |
| `token` | Token generated at [vPOS](https://merchant.vpos.ao) dashboard | `string`
| `posID` | Merchant POS ID provided by EMIS | `int64`
| `supervisorCard` | Merchant Supervisor Card number provided by EMIS | `string`
| `paymentCallbackURL` | Merchant application endpoint to accept the callback payment response | `string`
| `refundCallbackURL` | Merchant application endpoint to accept the callback refund response | `string`

#### Example
```go
package main

import (
    "os"

    "github.com/nellcorp/vpos-go"
)

var (
    posID              = os.Getenv("GPO_POS_ID")
    token              = os.Getenv("MERCHANT_VPOS_TOKEN")
    paymentCallbackURL = os.Getenv("PAYMENT_CALLBACK_URL")
    refundCallbackURL  = os.Getenv("REFUND_CALLBACK_URL")
    supervisorCard     = os.Getenv("GPO_SUPERVISOR_CARD")
)

func main() {
    vpos := vposgo.NewVPOS(
        posID,
        token,
        paymentCallbackURL,
        refundCallbackURL,
        supervisorCard
    )
}
```

### Payment Transaction
Creates a new payment transaction (or authorization request) given a `transactionType`, a valid `mobile number` associated with a `MULTICAIXA EXPRESS` account and a valid `amount`.

```go
package main

import (
    "log"
    "os"

    "github.com/nellcorp/vpos-go"
)

var (
    posID              = os.Getenv("GPO_POS_ID")
    token              = os.Getenv("MERCHANT_VPOS_TOKEN")
    paymentCallbackURL = os.Getenv("PAYMENT_CALLBACK_URL")
    refundCallbackURL  = os.Getenv("REFUND_CALLBACK_URL")
    supervisorCard     = os.Getenv("GPO_SUPERVISOR_CARD")
)

func main() {
    vpos := vposgo.NewVPOS(
        posID,
        token,
        paymentCallbackURL,
        refundCallbackURL,
        supervisorCard
    )

    transactionType := "payment" // or "authorization"
    mobile := "934174394"
    amount := "2000.00"

    transactionID, idempotencyKey, nonce, timeRemaining, err := vpos.PaymentTransaction(transactionType, mobile, amount)
    if err != nil {
		log.Fatal(err)
	}
}
```

| Argument | Description | Type |
| --- | --- | --- |
| `transactionType` | The transactionType is either 'payment' or 'authorization' | `string`
| `mobile` | The client mobile number in a string format. Must be a number registered in Multicaixa Express | `string`
| `amount` | Amount of the requested payment in a string format. Use (.) as decimals separator; at most 2 decimal places | `string`

### Payment Transaction with Authorization
Creates a new payment transaction from a previously accepted authorization transaction

```go
package main

import (
    "log"
    "os"

    "github.com/nellcorp/vpos-go"
)

var (
    posID              = os.Getenv("GPO_POS_ID")
    token              = os.Getenv("MERCHANT_VPOS_TOKEN")
    paymentCallbackURL = os.Getenv("PAYMENT_CALLBACK_URL")
    refundCallbackURL  = os.Getenv("REFUND_CALLBACK_URL")
    supervisorCard     = os.Getenv("GPO_SUPERVISOR_CARD")
)

func main() {
    vpos := vposgo.NewVPOS(
        posID,
        token,
        paymentCallbackURL,
        refundCallbackURL,
        supervisorCard
    )

    parentTransactionID := "9kOmKYUWxN0Jpe4PBoXzE"
    amount := "2000.00"

    transactionID, idempotencyKey, nonce, timeRemaining, err := vpos.PaymentWithAuthorization(parentTransactionID, amount)
    if err != nil {
		log.Fatal(err)
	}
}
```

| Argument | Description | Type |
| --- | --- | --- |
| `parentTransactionID` | The ID of the accepted authorization transaction | `string`
| `amount` | Amount to capture. Must be equal or less than the amount of the authorization. Use (.) as decimals separator; at most 2 decimal places | `string`

### Request Refund or Request Transaction Cancelation
Given an `transactionType` existing `parentTransactionID`, request a refund or transaction cancelation.

```go
package main

import (
    "log"
    "os"

    "github.com/nellcorp/vpos-go"
)

var (
    posID              = os.Getenv("GPO_POS_ID")
    token              = os.Getenv("MERCHANT_VPOS_TOKEN")
    paymentCallbackURL = os.Getenv("PAYMENT_CALLBACK_URL")
    refundCallbackURL  = os.Getenv("REFUND_CALLBACK_URL")
    supervisorCard     = os.Getenv("GPO_SUPERVISOR_CARD")
)

func main() {
    vpos := vposgo.NewVPOS(
        posID,
        token,
        paymentCallbackURL,
        refundCallbackURL,
        supervisorCard
    )

    transactionType := "refund" // or "cancelation"
    parentTransactionID := "9kOmKYUWxN0Jpe4PBoXzE"

    transactionID, idempotencyKey, nonce, timeRemaining, err := vpos.RefundOrCancelation(transactionType, parentTransactionID)
    if err != nil {
		log.Fatal(err)
	}
}
```

| Argument | Description | Type |
| --- | --- | --- |
| `transactionType` | The transactionType is either 'refund' or 'cancelation' | `string`
| `parentTransactionID` | This is a string value of the transaction id you're requesting to be refunded or canceled | `string`

### Get Transaction
This retrieves a specific transaction.

```go
package main

import (
    "log"
    "os"

    "github.com/nellcorp/vpos-go"
)

var (
    posID              = os.Getenv("GPO_POS_ID")
    token              = os.Getenv("MERCHANT_VPOS_TOKEN")
    paymentCallbackURL = os.Getenv("PAYMENT_CALLBACK_URL")
    refundCallbackURL  = os.Getenv("REFUND_CALLBACK_URL")
    supervisorCard     = os.Getenv("GPO_SUPERVISOR_CARD")
)

func main() {
    vpos := vposgo.NewVPOS(
        posID,
        token,
        paymentCallbackURL,
        refundCallbackURL,
        supervisorCard
    )

    transactionID := "1jHbXEbRTIbbwaoJ6w06nLcRG7X"
    transaction, err := vpos.GetTransaction(transactionID)
    if err != nil {
		log.Fatal(err)
	}

    /* 
        transaction = &Transaction{
            "amount": "101.09",
            "clearing_period": null,
            "id": "1jHbXEbRTIbbwaoJ6w06nLcRG7X",
            "mobile": "900111222",
            "parent_transaction_id": null,
            "pos_id": 100,
            "status": "accepted",
            "status_datetime": "2020-10-23T14:41:00Z",
            "status_reason": null,
            "type": "payment"
        }
    */
}
```

| Argument | Description | Type |
| --- | --- | --- |
| `transactionID` | The ID of the transaction to retrieve | `string`

We hope this module can be useful for you. Feel free to contribute and help to improve this package opening an `ISSUE` or `Pull Request` :grinning:.

License
----------------

The package is available as open source under the terms of the [MIT License](http://opensource.org/licenses/MIT).


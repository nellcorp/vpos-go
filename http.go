package vposgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	httpClient *http.Client

	statusReason = map[int64]string{
		1000: "Generic gateway error",
		1001: "Request timed-out and will not be processed",
		1002: "Gateway is not authorized to execute transactions on the specified POS",
		1003: "Parent transaction ID of refund request is not an accepted Payment",
		2000: "Generic processor error",
		2001: "Insufficient funds in client's account",
		2002: "Refused by the card issuer",
		2003: "Card or network daily limit exceeded",
		2004: "Request timed-out and was refused by the processor",
		2005: "POS is closed and unable to accept transactions",
		2006: "Insufficient funds in POS available for refund",
		2007: "Invalid or Inactive supervisor card",
		2008: "Invalid merchant email",
		2009: "Parent transaction is too old to be refunded",
		2010: "Request was refused by the processor",
		3000: "Refused by client",
	}
)

func init() {
	if httpClient == nil {
		setupHttpClient()
	}
}

func setupHttpClient() {
	httpClient = &http.Client{
		Timeout: time.Second * 15,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func httpGet(url string, headers map[string]string) (response []byte, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		err = fmt.Errorf("HTTP GET %s failed with: %s", url, err.Error())
		return
	}

	defer resp.Body.Close()

	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("HTTP GET %s failed with code %d", url, resp.StatusCode)
	}
	return
}

func httpPost(url string, headers map[string]string, params interface{}) (response []byte, responseHeaders http.Header, err error) {
	var payload []byte

	switch paramType := params.(type) {
	case nil:
		payload = []byte("")
	case []byte:
		payload = paramType
	case string:
		payload = []byte(paramType)
	default:
		payload, err = json.Marshal(params)
		if err != nil {
			return
		}
	}

	request, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = fmt.Errorf("HTTP POST Status: %s - Response: %s, Params: %v", resp.Status, string(response), params)
		return
	}
	responseHeaders = resp.Header.Clone()

	return
}

package nexmo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SMS represents the SMS API functions for sending text messages.
type SMS struct {
	client *Client
}

// SMS message types.
const (
	Text    = "text"
	Binary  = "binary"
	WAPPush = "wappush"
	Unicode = "unicode"
	VCal    = "vcal"
	VCard   = "vcard"
)

type MessageClass int

// SMS message classes.
const (
	// This type of SMS message is displayed on the mobile screen without being
	// saved in the message store or on the SIM card; unless explicitly saved
	// by the mobile user.
	Flash MessageClass = iota

	// This message is to be stored in the device memory or the SIM card
	// (depending on memory availability).
	Standard

	// This message class carries SIM card data. The SIM card data must be
	// successfully transferred prior to sending acknowledgment to the service
	// center. An error message is sent to the service center if this
	// transmission is not possible.
	SIMData

	// This message is forwarded from the receiving entity to an external
	// device. The delivery acknowledgment is sent to the service center
	// regardless of whether or not the message was forwarded to the external
	// device.
	Forward
)

var messageClassMap = map[MessageClass]string{
	Flash:    "flash",
	Standard: "standard",
	SIMData:  "SIM data",
	Forward:  "forward",
}

func (m MessageClass) String() string {
	return messageClassMap[m]
}
func (m *SMSMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ApiKey    string `json:"api_key"`
		ApiSecret string `json:"api_secret"`
		SMSMessage
	}{
		ApiKey:     m.apiKey,
		ApiSecret:  m.apiSecret,
		SMSMessage: *m,
	})
}

// Type SMSMessage defines a single SMS message.
type SMSMessage struct {
	apiKey               string
	apiSecret            string
	From                 string       `json:"from"`
	To                   string       `json:"to"`
	Type                 string       `json:"type"`
	Text                 string       `json:"text,omitempty"`              // Optional.
	StatusReportRequired int          `json:"status-report-req,omitempty"` // Optional.
	ClientReference      string       `json:"client-ref,omitempty"`        // Optional.
	NetworkCode          string       `json:"network-code,omitempty"`      // Optional.
	VCard                string       `json:"vcard,omitempty"`             // Optional.
	VCal                 string       `json:"vcal,omitempty"`              // Optional.
	TTL                  int          `json:"ttl,omitempty"`               // Optional.
	Class                MessageClass `json:"message-class,omitempty"`     // Optional.
	Body                 []byte       `json:"body,omitempty"`              // Required for Binary message.
	UDH                  []byte       `json:"udh,omitempty"`               // Required for Binary message.

	// The following is only for type=wappush

	Title    string `json:"title,omitempty"`    // Title shown to recipient
	URL      string `json:"url,omitempty"`      // WAP Push URL
	Validity int    `json:"validity,omitempty"` // Duration WAP Push is available in milliseconds
}

func (msg *SMSMessage) ToValues() url.Values {
	vals := url.Values{}
	vals.Add("from", msg.From)
	vals.Add("to", msg.To)
	vals.Add("type", msg.Type)
	if msg.Text != "" {
		vals.Add("text", msg.Text)
	}
	if msg.StatusReportRequired != 0 {
		vals.Add("status-report-req", strconv.Itoa(msg.StatusReportRequired))
	}
	if msg.ClientReference != "" {
		vals.Add("client-ref", msg.ClientReference)
	}
	if msg.NetworkCode != "" {
		vals.Add("network-code", msg.NetworkCode)
	}
	if msg.VCard != "" {
		vals.Add("vcard", msg.VCard)
	}
	if msg.VCal != "" {
		vals.Add("vcal", msg.VCal)
	}
	if msg.TTL != 0 {
		vals.Add("ttl", strconv.Itoa(msg.TTL))
	}
	// TODO support message-class
	if len(msg.Body) > 0 {
		vals.Add("body", string(msg.Body))
	}
	if len(msg.UDH) > 0 {
		vals.Add("udh", string(msg.UDH))
	}
	return vals
}

type ResponseCode int

func (c ResponseCode) String() string {
	return responseCodeMap[c]
}

const (
	ResponseSuccess ResponseCode = iota
	ResponseThrottled
	ResponseMissingParams
	ResponseInvalidParams
	ResponseInvalidCredentials
	ResponseInternalError
	ResponseInvalidMessage
	ResponseNumberBarred
	ResponsePartnerAcctBarred
	ResponsePartnerQuotaExceeded
	ResponseRESTNotEnabled
	ResponseMessageTooLong
	ResponseCommunicationFailed
	ResponseInvalidSignature
	ResponseInvalidSenderAddress
	ResponseInvalidTTL
	ResponseFacilityNotAllowed
	ResponseInvalidMessageClass
)

var responseCodeMap = map[ResponseCode]string{
	ResponseSuccess:              "Success",
	ResponseThrottled:            "Throttled",
	ResponseMissingParams:        "Missing params",
	ResponseInvalidParams:        "Invalid params",
	ResponseInvalidCredentials:   "Invalid credentials",
	ResponseInternalError:        "Internal error",
	ResponseInvalidMessage:       "Invalid message",
	ResponseNumberBarred:         "Number barred",
	ResponsePartnerAcctBarred:    "Partner account barred",
	ResponsePartnerQuotaExceeded: "Partner quota exceeded",
	ResponseRESTNotEnabled:       "Account not enabled for REST",
	ResponseMessageTooLong:       "Message too long",
	ResponseCommunicationFailed:  "Communication failed",
	ResponseInvalidSignature:     "Invalid signature",
	ResponseInvalidSenderAddress: "Invalid sender address",
	ResponseInvalidTTL:           "Invalid TTL",
	ResponseFacilityNotAllowed:   "Facility not allowed",
	ResponseInvalidMessageClass:  "Invalid message class",
}

// MessageReport is the "status report" for a single SMS sent via the Nexmo API
type MessageReport struct {
	Status           ResponseCode `json:"status,string"`
	MessageID        string       `json:"message-id"`
	To               string       `json:"to"`
	ClientReference  string       `json:"client-ref"`
	RemainingBalance string       `json:"remaining-balance"`
	MessagePrice     string       `json:"message-price"`
	Network          string       `json:"network"`
	ErrorText        string       `json:"error-text"`
}

// MessageResponse contains the response from Nexmo's API after we attempt to
// send any kind of message.
// It will contain one MessageReport for every 160 chars sent.
type MessageResponse struct {
	MessageCount int             `json:"message-count,string"`
	Messages     []MessageReport `json:"messages"`
}

// Send the message using the specified SMS client.
func (c *SMS) Send(msg *SMSMessage) (*MessageResponse, error) {
	if len(msg.From) <= 0 {
		return nil, errors.New("Invalid From field specified")
	}

	if len(msg.To) <= 0 {
		return nil, errors.New("Invalid To field specified")
	}

	if len(msg.ClientReference) > 40 {
		return nil, errors.New("Client reference too long")
	}

	var messageResponse *MessageResponse

	switch msg.Type {
	case Text:
	case Unicode:
		if len(msg.Text) <= 0 {
			return nil, errors.New("Invalid message text")
		}
	case Binary:
		if len(msg.UDH) == 0 || len(msg.Body) == 0 {
			return nil, errors.New("Invalid binary message")
		}

	case WAPPush:
		if len(msg.URL) == 0 || len(msg.Title) == 0 {
			return nil, errors.New("Invalid WAP Push parameters")
		}
	}
	if !c.client.useOauth {
		msg.apiKey = c.client.apiKey
		msg.apiSecret = c.client.apiSecret
	}

	client := &http.Client{}

	var r *http.Request

	messageValues := msg.ToValues()
	messageValues.Add("api_key", msg.apiKey)
	messageValues.Add("api_secret", msg.apiSecret)
	encodedForm := messageValues.Encode()
	if c.client.VerboseLogging {
		log.Println("NEXMO: Sending encoded form:", encodedForm)
	}
	r, _ = http.NewRequest("POST", apiRoot+"/sms/json", strings.NewReader(encodedForm))

	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if c.client.VerboseLogging {
		log.Printf("NEXMO: Sending request: %+v\n", r)
	}

	resp, err := client.Do(r)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.client.VerboseLogging {
		log.Println("NEXMO: Response status code:", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	if c.client.VerboseLogging {
		log.Println("NEXMO: Response:", string(body))
	}

	err = json.Unmarshal(body, &messageResponse)
	if err != nil {
		return nil, err
	}
	return messageResponse, nil
}

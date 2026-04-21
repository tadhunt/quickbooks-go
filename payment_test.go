package quickbooks

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPayment(t *testing.T) {
	jsonFile, err := os.Open("data/testing/payment.json")
	if err != nil {
		log.Fatal("When opening JSON file: ", err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal("When reading JSON file: ", err)
	}

	var r struct {
		Payment Payment
		Time    Date
	}
	err = json.Unmarshal(byteValue, &r)
	if err != nil {
		log.Fatal("When decoding JSON file: ", err)
	}

	assert.Equal(t, "3213", r.Payment.Id)
	assert.Equal(t, "0", r.Payment.SyncToken)
	assert.Equal(t, "QBO", r.Payment.Domain)
	assert.Equal(t, "434", r.Payment.CustomerRef.Value)
	assert.Equal(t, "Queens Public Library- Lefferts Branch - 9100", r.Payment.CustomerRef.Name)
	assert.Equal(t, "8", r.Payment.DepositToAccountRef.Value)
	assert.Equal(t, 300.00, r.Payment.TotalAmt)

	assertDateEqual(t, "2026-01-02", r.Payment.TxnDate)

	// New fields added to support manual-payment matching
	assert.Equal(t, "4", r.Payment.PaymentMethodRef.Value)
	assert.Equal(t, "211274457922958", r.Payment.PaymentRefNum)
	assert.Equal(t, "USD", r.Payment.CurrencyRef.Value)
	assert.Equal(t, "United States Dollar", r.Payment.CurrencyRef.Name)
	assert.Equal(t, 1.0, r.Payment.ExchangeRate)
	assert.True(t, strings.HasPrefix(r.Payment.PrivateNote, "ISA*00*NV"),
		"PrivateNote should preserve EDI 820 envelope, got: %q", r.Payment.PrivateNote)
	assert.Contains(t, r.Payment.PrivateNote, "RMR*OI*12/18/20259100",
		"PrivateNote should preserve embedded RMR remittance segment")

	// Line / linked transactions
	assert.Equal(t, 1, len(r.Payment.Line))
	assert.Equal(t, 300.00, r.Payment.Line[0].Amount)
	assert.Equal(t, 1, len(r.Payment.Line[0].LinkedTxn))
	assert.Equal(t, "2894", r.Payment.Line[0].LinkedTxn[0].TxnID)
	assert.Equal(t, "Invoice", r.Payment.Line[0].LinkedTxn[0].TxnType)

	assertDateEqual(t, "2026-01-07T08:34:14-08:00", r.Payment.MetaData.CreateTime)
	assertDateEqual(t, "2026-01-07T08:34:14-08:00", r.Payment.MetaData.LastUpdatedTime)

	// Round-trip: marshal back and verify the new fields survive serialization.
	out, err := json.Marshal(r.Payment)
	if err != nil {
		log.Fatal("When re-encoding payment: ", err)
	}
	roundTripped := string(out)
	assert.Contains(t, roundTripped, "\"PaymentRefNum\":\"211274457922958\"")
	assert.Contains(t, roundTripped, "\"PrivateNote\":")
	assert.Contains(t, roundTripped, "\"PaymentMethodRef\":")
	assert.Contains(t, roundTripped, "\"CurrencyRef\":")

	// Round-trip must preserve the bare date format on the wire (no
	// time component, no offset injected).
	assert.Contains(t, roundTripped, "\"TxnDate\":\"2026-01-02\"")
}

func TestNewDate(t *testing.T) {
	// Constructor produces a bare-date wire format suitable for outgoing
	// fields like Payment.TxnDate.
	tt := time.Date(2026, time.April, 21, 13, 45, 30, 0, time.UTC)
	d := NewDate(tt)

	out, err := json.Marshal(d)
	assert.NoError(t, err)
	assert.Equal(t, `"2026-04-21"`, string(out))
}

func TestNewDateTime(t *testing.T) {
	// Constructor produces an RFC3339 wire format suitable for outgoing
	// fields like MetaData.CreateTime.
	tt := time.Date(2026, time.April, 21, 13, 45, 30, 0, time.UTC)
	d := NewDateTime(tt)

	out, err := json.Marshal(d)
	assert.NoError(t, err)
	assert.Equal(t, `"2026-04-21T13:45:30Z"`, string(out))
}

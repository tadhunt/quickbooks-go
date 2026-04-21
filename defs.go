// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type CustomField struct {
	DefinitionId string `json:"DefinitionId,omitempty"`
	StringValue  string `json:"StringValue,omitempty"`
	Type         string `json:"Type,omitempty"`
	Name         string `json:"Name,omitempty"`
}

// Date holds a QuickBooks date or timestamp as the raw JSON bytes
// returned by the API. QuickBooks returns two distinct shapes:
//
//   - A bare calendar date such as "2026-01-02" (used for fields like
//     TxnDate, DueDate, CompanyStartDate). These have no timezone in
//     the wire format and are intended to be interpreted in the
//     QuickBooks company's configured timezone.
//
//   - An RFC3339 timestamp such as "2026-01-07T08:34:14-08:00" (used
//     for MetaData.CreateTime, MetaData.LastUpdatedTime, etc.). These
//     carry their own offset.
//
// Date deliberately does not parse on UnmarshalJSON because the
// standard library's encoding/json passes no context to the
// unmarshaler and bare dates can only be correctly anchored when
// the company timezone is known. Parse with In(loc) at the call
// site instead. For convenience, Client.Time(d) parses against the
// company timezone fetched at Client construction.
type Date struct {
	json.RawMessage
}

// In returns the Date interpreted in loc. RFC3339 strings carry their
// own offset and ignore loc; bare YYYY-MM-DD strings are anchored at
// midnight in loc. An empty or null Date returns the zero time.
func (d Date) In(loc *time.Location) (time.Time, error) {
	if len(d.RawMessage) == 0 || bytes.Equal(d.RawMessage, []byte("null")) {
		return time.Time{}, nil
	}
	s := string(bytes.Trim(d.RawMessage, `"`))
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, loc); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("quickbooks: unrecognized date %q", s)
}

// NewDate constructs an outgoing bare date from t (YYYY-MM-DD).
// Use for fields like TxnDate and DueDate.
func NewDate(t time.Time) Date {
	return Date{RawMessage: []byte(fmt.Sprintf("%q", t.Format("2006-01-02")))}
}

// NewDateTime constructs an outgoing RFC3339 timestamp from t.
// Use for fields like MetaData.CreateTime where a full datetime is expected.
func NewDateTime(t time.Time) Date {
	return Date{RawMessage: []byte(fmt.Sprintf("%q", t.Format(time.RFC3339)))}
}

// EmailAddress represents a QuickBooks email address.
type EmailAddress struct {
	Address string `json:",omitempty"`
}

// EndpointUrl specifies the endpoint to connect to
type EndpointUrl string

const (
	// DiscoveryProductionEndpoint is for live apps.
	DiscoveryProductionEndpoint EndpointUrl = "https://developer.api.intuit.com/.well-known/openid_configuration"
	// DiscoverySandboxEndpoint is for testing.
	DiscoverySandboxEndpoint EndpointUrl = "https://developer.api.intuit.com/.well-known/openid_sandbox_configuration"
	// ProductionEndpoint is for live apps.
	ProductionEndpoint EndpointUrl = "https://quickbooks.api.intuit.com"
	// SandboxEndpoint is for testing.
	SandboxEndpoint EndpointUrl = "https://sandbox-quickbooks.api.intuit.com"

	queryPageSize = 1000
)

func (u EndpointUrl) String() string {
	return string(u)
}

// MemoRef represents a QuickBooks MemoRef object.
type MemoRef struct {
	Value string `json:"value,omitempty"`
}

// MetaData is a timestamp of genesis and last change of a Quickbooks object
type MetaData struct {
	CreateTime      Date `json:",omitempty"`
	LastUpdatedTime Date `json:",omitempty"`
}

// PhysicalAddress represents a QuickBooks address.
type PhysicalAddress struct {
	Id string `json:"Id,omitempty"`
	// These lines are context-dependent! Read the QuickBooks API carefully.
	Line1   string `json:",omitempty"`
	Line2   string `json:",omitempty"`
	Line3   string `json:",omitempty"`
	Line4   string `json:",omitempty"`
	Line5   string `json:",omitempty"`
	City    string `json:",omitempty"`
	Country string `json:",omitempty"`
	// A.K.A. State.
	CountrySubDivisionCode string `json:",omitempty"`
	PostalCode             string `json:",omitempty"`
	Lat                    string `json:",omitempty"`
	Long                   string `json:",omitempty"`
}

// ReferenceType represents a QuickBooks reference to another object.
type ReferenceType struct {
	Value string `json:"value,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
}

// TelephoneNumber represents a QuickBooks phone number.
type TelephoneNumber struct {
	FreeFormNumber string `json:",omitempty"`
}

// WebSiteAddress represents a Quickbooks Website
type WebSiteAddress struct {
	URI string `json:",omitempty"`
}

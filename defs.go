// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"time"

	"github.com/markusmobius/go-dateparser"
)

type CustomField struct {
	DefinitionId string `json:"DefinitionId,omitempty"`
	StringValue  string `json:"StringValue,omitempty"`
	Type         string `json:"Type,omitempty"`
	Name         string `json:"Name,omitempty"`
}

// Date represents a Quickbooks date
type Date struct {
	time.Time `json:",omitempty"`
	raw  []byte
}

func (d *Date) UnmarshalJSON(b []byte) error {
	d.raw = make([]byte, len(b))
	copy(d.raw, b)

	if len(b) == 0 {
		d.Time = time.Time{}
		return nil
	}

	if len(b) > 1 &&  b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	dpcfg := &dateparser.Configuration{
		DefaultTimezone: time.Local,
	}

	date, err := dateparser.Parse(dpcfg, string(b))
	if err != nil {
		return err
	}

	d.Time = date.Time

	return nil
}

func (d Date) String() string {
	return d.Format(DateFormat)
}

func (d Date) GetRaw() []byte {
	return d.raw
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

	DateFormat        = "2006-01-02T15:04:05-07:00"
	queryPageSize = 1000
//	secondFormat  = "2006-01-02"
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

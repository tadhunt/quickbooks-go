package quickbooks

import (
	"fmt"
	"encoding/json"
)

// Vendor describes a vendor.
type Vendor struct {
	Id               string       `json:"Id,omitempty"`
	SyncToken        string       `json:",omitempty"`
	Title            string       `json:",omitempty"`
	GivenName        string       `json:",omitempty"`
	MiddleName       string       `json:",omitempty"`
	Suffix           string       `json:",omitempty"`
	FamilyName       string       `json:",omitempty"`
	PrimaryEmailAddr EmailAddress `json:",omitempty"`
	DisplayName      string       `json:",omitempty"`
	// ContactInfo
	APAccountRef      ReferenceType   `json:",omitempty"`
	TermRef           ReferenceType   `json:",omitempty"`
	GSTIN             string          `json:",omitempty"`
	Fax               TelephoneNumber `json:",omitempty"`
	BusinessNumber    string          `json:",omitempty"`
	CurrencyRef       ReferenceType   `json:",omitempty"`
	HasTPAR           bool            `json:",omitempty"`
	TaxReportingBasis string          `json:",omitempty"`
	Mobile            TelephoneNumber `json:",omitempty"`
	PrimaryPhone      TelephoneNumber `json:",omitempty"`
	Active            bool            `json:",omitempty"`
	AlternatePhone    TelephoneNumber `json:",omitempty"`
	MetaData          MetaData        `json:",omitempty"`
	Vendor1099        bool            `json:",omitempty"`
	BillRate          json.Number     `json:",omitempty"`
	WebAddr           *WebSiteAddress `json:",omitempty"`
	CompanyName       string          `json:",omitempty"`
	// VendorPaymentBankDetail
	TaxIdentifier       string           `json:",omitempty"`
	AcctNum             string           `json:",omitempty"`
	GSTRegistrationType string           `json:",omitempty"`
	PrintOnCheckName    string           `json:",omitempty"`
	BillAddr            *PhysicalAddress `json:",omitempty"`
	Balance             json.Number      `json:",omitempty"`
}

// CreateVendor creates the given Vendor on the QuickBooks server, returning
// the resulting Vendor object.
func (c *Client) CreateVendor(vendor *Vendor) (*Vendor, error) {
	var resp struct {
		Vendor Vendor
		Time   Date
	}

	if err := c.post("vendor", vendor, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Vendor, nil
}

// EntityName returns the QuickBooks entity name for Vendor.
func (Vendor) EntityName() string { return "Vendor" }

func (v Vendor) entityId() string { return v.Id }

func (v Vendor) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindVendors gets the full list of Vendors in the QuickBooks account.
func (c *Client) FindVendors() ([]Vendor, error) {
	return findAll[Vendor](c)
}

// FindVendorById finds the vendor by the given id
func (c *Client) FindVendorById(id string) (*Vendor, error) {
	var resp struct {
		Vendor Vendor
		Time   Date
	}

	if err := c.get("vendor/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Vendor, nil
}

// UpdateVendor updates the vendor
func (c *Client) UpdateVendor(vendor *Vendor) (*Vendor, error) {
	if vendor.Id == "" {
		return nil, fmt.Errorf("%w: missing vendor id", ErrMissingID)
	}

	existingVendor, err := c.FindVendorById(vendor.Id)
	if err != nil {
		return nil, err
	}

	vendor.SyncToken = existingVendor.SyncToken

	payload := struct {
		*Vendor
		Sparse bool `json:"sparse"`
	}{
		Vendor: vendor,
		Sparse: true,
	}

	var vendorData struct {
		Vendor Vendor
		Time   Date
	}

	if err = c.post("vendor", payload, &vendorData, nil); err != nil {
		return nil, err
	}

	return &vendorData.Vendor, err
}

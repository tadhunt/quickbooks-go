// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

// PaymentMethod represents a QuickBooks PaymentMethod entity (the display-name
// table referenced by Payment.PaymentMethodRef). QB Online exposes only the
// ID on read-side references, so fetching the full list is required to
// resolve IDs like "4" into names like "Check" or "Visa".
//
// See https://developer.intuit.com/app/developer/qbapi-docs/api/accounting/all-entities/paymentmethod
type PaymentMethod struct {
	Id        string   `json:"Id,omitempty"`
	SyncToken string   `json:",omitempty"`
	MetaData  MetaData `json:",omitempty"`
	Domain    string   `json:"domain,omitempty"`
	Name      string
	Active    bool   `json:",omitempty"`
	Type      string `json:",omitempty"` // "CREDIT_CARD" or "NON_CREDIT_CARD"
}

// CreatePaymentMethod creates a PaymentMethod in QuickBooks.
func (c *Client) CreatePaymentMethod(pm *PaymentMethod) (*PaymentMethod, error) {
	var resp struct {
		PaymentMethod PaymentMethod
		Time          Date
	}

	if err := c.post("paymentmethod", pm, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.PaymentMethod, nil
}

// EntityName returns the QuickBooks entity name for PaymentMethod.
func (PaymentMethod) EntityName() string { return "PaymentMethod" }

func (v PaymentMethod) entityId() string { return v.Id }

func (v PaymentMethod) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindPaymentMethods gets the full list of PaymentMethods in the QuickBooks account.
func (c *Client) FindPaymentMethods() ([]PaymentMethod, error) {
	return findAll[PaymentMethod](c)
}

// FindPaymentMethodById returns a PaymentMethod with a given Id.
func (c *Client) FindPaymentMethodById(id string) (*PaymentMethod, error) {
	var resp struct {
		PaymentMethod PaymentMethod
		Time          Date
	}

	if err := c.get("paymentmethod/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.PaymentMethod, nil
}

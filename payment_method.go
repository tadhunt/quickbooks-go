// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"fmt"
	"strconv"
)

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

// FindPaymentMethods gets the full list of PaymentMethods in the QuickBooks
// realm. Paginated in case a realm has many entries (most have < 20).
func (c *Client) FindPaymentMethods() ([]PaymentMethod, error) {
	var resp struct {
		QueryResponse struct {
			PaymentMethods []PaymentMethod `json:"PaymentMethod"`
			MaxResults     int
			StartPosition  int
			TotalCount     int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM PaymentMethod", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, fmt.Errorf("%w: no payment methods could be found", ErrNotFound)
	}

	paymentMethods := make([]PaymentMethod, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += queryPageSize {
		query := "SELECT * FROM PaymentMethod ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(queryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.PaymentMethods == nil {
			return nil, fmt.Errorf("%w: no payment methods could be found", ErrNotFound)
		}

		paymentMethods = append(paymentMethods, resp.QueryResponse.PaymentMethods...)
	}

	return paymentMethods, nil
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

// QueryPaymentMethods accepts an SQL query and returns all PaymentMethods
// found using it.
func (c *Client) QueryPaymentMethods(query string) ([]PaymentMethod, error) {
	var resp struct {
		QueryResponse struct {
			PaymentMethods []PaymentMethod `json:"PaymentMethod"`
			StartPosition  int
			MaxResults     int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.PaymentMethods == nil {
		return nil, fmt.Errorf("%w: could not find any payment methods", ErrNotFound)
	}

	return resp.QueryResponse.PaymentMethods, nil
}

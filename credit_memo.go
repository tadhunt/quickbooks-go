package quickbooks

import (
	"fmt"
	"encoding/json"
)

type CreditMemo struct {
	TotalAmt              float64         `json:",omitempty"`
	RemainingCredit       json.Number     `json:",omitempty"`
	Line                  []Line          `json:",omitempty"`
	ApplyTaxAfterDiscount bool            `json:",omitempty"`
	DocNumber             string          `json:",omitempty"`
	TxnDate               Date            `json:",omitempty"`
	Sparse                bool            `json:"sparse,omitempty"`
	CustomerMemo          MemoRef         `json:",omitempty"`
	ProjectRef            ReferenceType   `json:",omitempty"`
	Balance               json.Number     `json:",omitempty"`
	CustomerRef           ReferenceType   `json:",omitempty"`
	TxnTaxDetail          *TxnTaxDetail   `json:",omitempty"`
	SyncToken             string          `json:",omitempty"`
	CustomField           []CustomField   `json:",omitempty"`
	ShipAddr              PhysicalAddress `json:",omitempty"`
	EmailStatus           string          `json:",omitempty"`
	BillAddr              PhysicalAddress `json:",omitempty"`
	MetaData              MetaData        `json:",omitempty"`
	BillEmail             EmailAddress    `json:",omitempty"`
	Id                    string          `json:",omitempty"`
}

// CreateCreditMemo creates the given CreditMemo witin QuickBooks.
func (c *Client) CreateCreditMemo(creditMemo *CreditMemo) (*CreditMemo, error) {
	var resp struct {
		CreditMemo CreditMemo
		Time       Date
	}

	if err := c.post("creditmemo", creditMemo, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.CreditMemo, nil
}

// DeleteCreditMemo deletes the given credit memo.
func (c *Client) DeleteCreditMemo(creditMemo *CreditMemo) error {
	if creditMemo.Id == "" || creditMemo.SyncToken == "" {
		return fmt.Errorf("%w: missing id/sync token", ErrMissingID)
	}

	return c.post("creditmemo", creditMemo, nil, map[string]string{"operation": "delete"})
}

// EntityName returns the QuickBooks entity name for CreditMemo.
func (CreditMemo) EntityName() string { return "CreditMemo" }

func (v CreditMemo) entityId() string { return v.Id }

func (v CreditMemo) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindCreditMemos gets the full list of CreditMemos in the QuickBooks account.
func (c *Client) FindCreditMemos() ([]CreditMemo, error) {
	return findAll[CreditMemo](c)
}

// FindCreditMemoById retrieves the given credit memo from QuickBooks.
func (c *Client) FindCreditMemoById(id string) (*CreditMemo, error) {
	var resp struct {
		CreditMemo CreditMemo
		Time       Date
	}

	if err := c.get("creditmemo/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.CreditMemo, nil
}

// UpdateCreditMemo updates the given credit memo.
func (c *Client) UpdateCreditMemo(creditMemo *CreditMemo) (*CreditMemo, error) {
	if creditMemo.Id == "" {
		return nil, fmt.Errorf("%w: missing credit memo id", ErrMissingID)
	}

	existingCreditMemo, err := c.FindCreditMemoById(creditMemo.Id)
	if err != nil {
		return nil, err
	}

	creditMemo.SyncToken = existingCreditMemo.SyncToken

	payload := struct {
		*CreditMemo
		Sparse bool `json:"sparse"`
	}{
		CreditMemo: creditMemo,
		Sparse:     true,
	}

	var creditMemoData struct {
		CreditMemo CreditMemo
		Time       Date
	}

	if err = c.post("creditmemo", payload, &creditMemoData, nil); err != nil {
		return nil, err
	}

	return &creditMemoData.CreditMemo, err
}

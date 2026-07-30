package quickbooks

import (
	"fmt"
	"encoding/json"
)

type Bill struct {
	Id           string        `json:"Id,omitempty"`
	VendorRef    ReferenceType `json:",omitempty"`
	Line         []Line
	SyncToken    string        `json:",omitempty"`
	CurrencyRef  ReferenceType `json:",omitempty"`
	TxnDate      Date          `json:",omitempty"`
	APAccountRef ReferenceType `json:",omitempty"`
	SalesTermRef ReferenceType `json:",omitempty"`
	LinkedTxn    []LinkedTxn   `json:",omitempty"`
	// GlobalTaxCalculation
	TotalAmt                json.Number `json:",omitempty"`
	TransactionLocationType string      `json:",omitempty"`
	DueDate                 Date        `json:",omitempty"`
	MetaData                MetaData    `json:",omitempty"`
	DocNumber               string
	PrivateNote             string        `json:",omitempty"`
	TxnTaxDetail            TxnTaxDetail  `json:",omitempty"`
	ExchangeRate            json.Number   `json:",omitempty"`
	DepartmentRef           ReferenceType `json:",omitempty"`
	IncludeInAnnualTPAR     bool          `json:",omitempty"`
	HomeBalance             json.Number   `json:",omitempty"`
	RecurDataRef            ReferenceType `json:",omitempty"`
	Balance                 json.Number   `json:",omitempty"`
}

// CreateBill creates the given Bill on the QuickBooks server, returning
// the resulting Bill object.
func (c *Client) CreateBill(bill *Bill) (*Bill, error) {
	var resp struct {
		Bill Bill
		Time Date
	}

	if err := c.post("bill", bill, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Bill, nil
}

// DeleteBill deletes the bill
func (c *Client) DeleteBill(bill *Bill) error {
	if bill.Id == "" || bill.SyncToken == "" {
		return fmt.Errorf("%w: missing id/sync token", ErrMissingID)
	}

	return c.post("bill", bill, nil, map[string]string{"operation": "delete"})
}

// EntityName returns the QuickBooks entity name for Bill.
func (Bill) EntityName() string { return "Bill" }

func (v Bill) entityId() string { return v.Id }

func (v Bill) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindBills gets the full list of Bills in the QuickBooks account.
func (c *Client) FindBills() ([]Bill, error) {
	return findAll[Bill](c)
}

// FindBillById finds the bill by the given id
func (c *Client) FindBillById(id string) (*Bill, error) {
	var resp struct {
		Bill Bill
		Time Date
	}

	if err := c.get("bill/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Bill, nil
}

// QueryBills accepts an SQL query and returns all bills found using it
func (c *Client) QueryBills(query string) ([]Bill, error) {
	var resp struct {
		QueryResponse struct {
			Bills         []Bill `json:"Bill"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Bills == nil {
		return nil, fmt.Errorf("%w: could not find any bills", ErrNotFound)
	}

	return resp.QueryResponse.Bills, nil
}

// UpdateBill updates the bill
func (c *Client) UpdateBill(bill *Bill) (*Bill, error) {
	if bill.Id == "" {
		return nil, fmt.Errorf("%w: missing bill id", ErrMissingID)
	}

	existingBill, err := c.FindBillById(bill.Id)
	if err != nil {
		return nil, err
	}

	bill.SyncToken = existingBill.SyncToken

	payload := struct {
		*Bill
		Sparse bool `json:"sparse"`
	}{
		Bill:   bill,
		Sparse: true,
	}

	var billData struct {
		Bill Bill
		Time Date
	}

	if err = c.post("bill", payload, &billData, nil); err != nil {
		return nil, err
	}

	return &billData.Bill, err
}

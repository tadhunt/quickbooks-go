package quickbooks

import (
	"fmt"
	"errors"
)

type Deposit struct {
	SyncToken           string        `json:",omitempty"`
	Domain              string        `json:"domain,omitempty"`
	DepositToAccountRef ReferenceType `json:",omitempty"`
	TxnDate             Date          `json:",omitempty"`
	TotalAmt            float64       `json:",omitempty"`
	Line                []PaymentLine `json:",omitempty"`
	Id                  string        `json:",omitempty"`
	MetaData            MetaData      `json:",omitempty"`
}

// CreateDeposit creates the given deposit within QuickBooks
func (c *Client) CreateDeposit(deposit *Deposit) (*Deposit, error) {
	var resp struct {
		Deposit Deposit
		Time    Date
	}

	if err := c.post("deposit", deposit, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Deposit, nil
}

func (c *Client) DeleteDeposit(deposit *Deposit) error {
	if deposit.Id == "" || deposit.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post("deposit", deposit, nil, map[string]string{"operation": "delete"})
}

// EntityName returns the QuickBooks entity name for Deposit.
func (Deposit) EntityName() string { return "Deposit" }

func (v Deposit) entityId() string { return v.Id }

func (v Deposit) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindDeposits gets the full list of Deposits in the QuickBooks account.
func (c *Client) FindDeposits() ([]Deposit, error) {
	return findAll[Deposit](c)
}

// FindDepositById returns an deposit with a given Id.
func (c *Client) FindDepositById(id string) (*Deposit, error) {
	var resp struct {
		Deposit Deposit
		Time    Date
	}

	if err := c.get("deposit/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Deposit, nil
}

// UpdateDeposit updates the deposit
func (c *Client) UpdateDeposit(deposit *Deposit) (*Deposit, error) {
	if deposit.Id == "" {
		return nil, fmt.Errorf("%w: missing deposit id", ErrMissingID)
	}

	existingDeposit, err := c.FindDepositById(deposit.Id)
	if err != nil {
		return nil, err
	}

	deposit.SyncToken = existingDeposit.SyncToken

	payload := struct {
		*Deposit
		Sparse bool `json:"sparse"`
	}{
		Deposit: deposit,
		Sparse:  true,
	}

	var depositData struct {
		Deposit Deposit
		Time    Date
	}

	if err = c.post("deposit", payload, &depositData, nil); err != nil {
		return nil, err
	}

	return &depositData.Deposit, err
}

package quickbooks

import (
	"fmt"
)

type TimeActivity struct {
	Id             string         `json:",omitempty"`
	SyncToken      string         `json:",omitempty"`
	MetaData       MetaData       `json:",omitempty"`
	TxnDate        string         `json:"TxnDate,omitempty"`
	NameOf         string         `json:"NameOf,omitempty"` // "Employee" or "Vendor"
	EmployeeRef    *ReferenceType `json:"EmployeeRef,omitempty"`
	ItemRef        *ReferenceType `json:"ItemRef,omitempty"` // Service item (e.g. "Sales Commission"); nil for regular hours
	Hours          int            `json:"Hours,omitempty"`
	Minutes        int            `json:"Minutes,omitempty"`
	Description    string         `json:"Description,omitempty"`
	BillableStatus string         `json:"BillableStatus,omitempty"`
}

// CreateTimeActivity creates the given time activity within QuickBooks
func (c *Client) CreateTimeActivity(timeActivity *TimeActivity) (*TimeActivity, error) {
	var resp struct {
		TimeActivity TimeActivity
		Time         Date
	}

	if err := c.post("timeactivity", timeActivity, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TimeActivity, nil
}

// EntityName returns the QuickBooks entity name for TimeActivity.
func (TimeActivity) EntityName() string { return "TimeActivity" }

func (v TimeActivity) entityId() string { return v.Id }

func (v TimeActivity) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindTimeActivities gets the full list of TimeActivities in the QuickBooks account.
func (c *Client) FindTimeActivities() ([]TimeActivity, error) {
	return findAll[TimeActivity](c)
}

// FindTimeActivityById returns a time activity with a given Id.
func (c *Client) FindTimeActivityById(id string) (*TimeActivity, error) {
	var resp struct {
		TimeActivity TimeActivity
		Time         Date
	}

	if err := c.get("timeactivity/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TimeActivity, nil
}

// QueryTimeActivities accepts an SQL query and returns all time activities found using it
func (c *Client) QueryTimeActivities(query string) ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivity  []TimeActivity `json:"TimeActivity"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TimeActivity == nil {
		return nil, fmt.Errorf("%w: could not find any time activities", ErrNotFound)
	}

	return resp.QueryResponse.TimeActivity, nil
}

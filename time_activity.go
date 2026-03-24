package quickbooks

import (
	"errors"
	"strconv"
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

// FindTimeActivities gets the full list of TimeActivities in the QuickBooks account.
func (c *Client) FindTimeActivities() ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivity  []TimeActivity `json:"TimeActivity"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM TimeActivity", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no time activities could be found")
	}

	timeActivities := make([]TimeActivity, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += queryPageSize {
		query := "SELECT * FROM TimeActivity ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(queryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.TimeActivity == nil {
			return nil, errors.New("no time activities could be found")
		}

		timeActivities = append(timeActivities, resp.QueryResponse.TimeActivity...)
	}

	return timeActivities, nil
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
		return nil, errors.New("could not find any time activities")
	}

	return resp.QueryResponse.TimeActivity, nil
}

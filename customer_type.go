package quickbooks

type CustomerType struct {
	SyncToken string   `json:",omitempty"`
	Domain    string   `json:"domain,omitempty"`
	Name      string   `json:",omitempty"`
	Active    bool     `json:",omitempty"`
	Id        string   `json:",omitempty"`
	MetaData  MetaData `json:",omitempty"`
}

// FindCustomerTypeById returns a customerType with a given Id.
func (c *Client) FindCustomerTypeById(id string) (*CustomerType, error) {
	var r struct {
		CustomerType CustomerType
		Time         Date
	}

	if err := c.get("customertype/"+id, &r, nil); err != nil {
		return nil, err
	}

	return &r.CustomerType, nil
}

// EntityName returns the QuickBooks entity name for CustomerType.
func (CustomerType) EntityName() string { return "CustomerType" }

func (v CustomerType) entityId() string { return v.Id }

func (v CustomerType) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindCustomerTypes gets the full list of CustomerTypes in the QuickBooks account.
func (c *Client) FindCustomerTypes() ([]CustomerType, error) {
	return findAll[CustomerType](c)
}

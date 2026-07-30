package quickbooks

import (
	"fmt"
)

type Employee struct {
	SyncToken        string          `json:",omitempty"`
	Domain           string          `json:"domain,omitempty"`
	DisplayName      string          `json:",omitempty"`
	PrimaryPhone     TelephoneNumber `json:",omitempty"`
	PrintOnCheckName string          `json:",omitempty"`
	FamilyName       string          `json:",omitempty"`
	Active           bool            `json:",omitempty"`
	SSN              string          `json:",omitempty"`
	PrimaryAddr      PhysicalAddress `json:",omitempty"`
	PrimaryEmailAddr EmailAddress    `json:",omitempty"`
	BillableTime     bool            `json:",omitempty"`
	GivenName        string          `json:",omitempty"`
	Id               string          `json:",omitempty"`
	MetaData         MetaData        `json:",omitempty"`
}

// CreateEmployee creates the given employee within QuickBooks
func (c *Client) CreateEmployee(employee *Employee) (*Employee, error) {
	var resp struct {
		Employee Employee
		Time     Date
	}

	if err := c.post("employee", employee, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

// EntityName returns the QuickBooks entity name for Employee.
func (Employee) EntityName() string { return "Employee" }

func (v Employee) entityId() string { return v.Id }

func (v Employee) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindEmployees gets the full list of Employees in the QuickBooks account.
func (c *Client) FindEmployees() ([]Employee, error) {
	return findAll[Employee](c)
}

// FindEmployeeById returns an employee with a given Id.
func (c *Client) FindEmployeeById(id string) (*Employee, error) {
	var resp struct {
		Employee Employee
		Time     Date
	}

	if err := c.get("employee/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

// UpdateEmployee updates the employee
func (c *Client) UpdateEmployee(employee *Employee) (*Employee, error) {
	if employee.Id == "" {
		return nil, fmt.Errorf("%w: missing employee id", ErrMissingID)
	}

	existingEmployee, err := c.FindEmployeeById(employee.Id)
	if err != nil {
		return nil, err
	}

	employee.SyncToken = existingEmployee.SyncToken

	payload := struct {
		*Employee
		Sparse bool `json:"sparse"`
	}{
		Employee: employee,
		Sparse:   true,
	}

	var employeeData struct {
		Employee Employee
		Time     Date
	}

	if err = c.post("employee", payload, &employeeData, nil); err != nil {
		return nil, err
	}

	return &employeeData.Employee, err
}

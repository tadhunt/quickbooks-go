// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"fmt"
	"encoding/json"
)

// Item represents a QuickBooks Item object (a product type).
type Item struct {
	Id          string   `json:"Id,omitempty"`
	SyncToken   string   `json:",omitempty"`
	MetaData    MetaData `json:",omitempty"`
	Name        string
	SKU         string `json:"Sku,omitempty"`
	Description string `json:",omitempty"`
	Active      bool   `json:",omitempty"`
	// SubItem
	// ParentRef
	// Level
	// FullyQualifiedName
	Taxable             bool        `json:",omitempty"`
	SalesTaxIncluded    bool        `json:",omitempty"`
	UnitPrice           json.Number `json:",omitempty"`
	Type                string
	IncomeAccountRef    ReferenceType
	ExpenseAccountRef   ReferenceType
	PurchaseDesc        string      `json:",omitempty"`
	PurchaseTaxIncluded bool        `json:",omitempty"`
	PurchaseCost        json.Number `json:",omitempty"`
	AssetAccountRef     ReferenceType
	TrackQtyOnHand      bool `json:",omitempty"`
	// InvStartDate Date
	QtyOnHand          json.Number   `json:",omitempty"`
	SalesTaxCodeRef    ReferenceType `json:",omitempty"`
	PurchaseTaxCodeRef ReferenceType `json:",omitempty"`
}

func (c *Client) CreateItem(item *Item) (*Item, error) {
	var resp struct {
		Item Item
		Time Date
	}

	if err := c.post("item", item, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// EntityName returns the QuickBooks entity name for Item.
func (Item) EntityName() string { return "Item" }

func (v Item) entityId() string { return v.Id }

func (v Item) entityCreateTime() Date { return v.MetaData.CreateTime }

// FindItems gets the full list of Items in the QuickBooks account.
func (c *Client) FindItems() ([]Item, error) {
	return findAll[Item](c)
}

// FindItemById returns an item with a given Id.
func (c *Client) FindItemById(id string) (*Item, error) {
	var resp struct {
		Item Item
		Time Date
	}

	if err := c.get("item/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// UpdateItem updates the item
func (c *Client) UpdateItem(item *Item) (*Item, error) {
	if item.Id == "" {
		return nil, fmt.Errorf("%w: missing item id", ErrMissingID)
	}

	existingItem, err := c.FindItemById(item.Id)
	if err != nil {
		return nil, err
	}

	item.SyncToken = existingItem.SyncToken

	payload := struct {
		*Item
		Sparse bool `json:"sparse"`
	}{
		Item:   item,
		Sparse: true,
	}

	var itemData struct {
		Item Item
		Time Date
	}

	if err = c.post("item", payload, &itemData, nil); err != nil {
		return nil, err
	}

	return &itemData.Item, err
}

package entities

import "time"

type ID uint64

type InventoryData struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	FullPriceTHB   uint64 `json:"fullPriceTHB"`
	Count          uint64 `json:"count"`
	ManufacturerID uint64 `json:"manufacturerId"`
}

type PartialInventoryData struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	FullPriceTHB   *uint64 `json:"fullPriceTHB"`
	Count          *uint64 `json:"count"`
	ManufacturerID *uint64 `json:"manufacturerId"`
}

type Inventory struct {
	ID ID `json:"id"`
	InventoryData
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type GetInventoriesFilter struct {
	Cursor ID
	Limit  uint64
}

type PaginatedInventories struct {
	Inventories []Inventory
	Cursor      ID
}

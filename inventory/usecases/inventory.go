package usecases

import (
	"context"
	"fmt"
	"go-http-serve/inventory/entities"
	"go-http-serve/inventory/repositories"
)

type InventoryCRUDUseCases interface {
	CreateInventory(ctx context.Context, inventory entities.InventoryData) (entities.ID, error)
	GetInventories(ctx context.Context, filter entities.GetInventoriesFilter) (entities.PaginatedInventories, error)
	UpdateInventory(ctx context.Context, id entities.ID, partialInventory entities.PartialInventoryData) error
	DeleteInventory(ctx context.Context, id entities.ID) error
}

type inventoryCRUDUseCasesImpl struct {
	inventoryRepository repositories.InventoryRepository
}

func NewInventoryCRUDUseCases(inventoryRepository repositories.InventoryRepository) InventoryCRUDUseCases {
	return &inventoryCRUDUseCasesImpl{
		inventoryRepository: inventoryRepository,
	}
}

func (u *inventoryCRUDUseCasesImpl) CreateInventory(ctx context.Context, inventory entities.InventoryData) (entities.ID, error) {
	id, err := u.inventoryRepository.Create(ctx, inventory)
	if err != nil {
		return id, fmt.Errorf("unable to create inventory: %w", err)
	}

	return id, nil
}

func (u *inventoryCRUDUseCasesImpl) GetInventories(ctx context.Context, filter entities.GetInventoriesFilter) (entities.PaginatedInventories, error) {
	inventories, err := u.inventoryRepository.Find(ctx, filter)
	if err != nil {
		return inventories, fmt.Errorf("unable to get inventories: %w", err)
	}

	return inventories, nil
}

func (u *inventoryCRUDUseCasesImpl) UpdateInventory(ctx context.Context, id entities.ID, partialInventory entities.PartialInventoryData) error {
	if err := u.inventoryRepository.Update(ctx, id, partialInventory); err != nil {
		return fmt.Errorf("unable to update inventory: %w", err)
	}

	return nil
}

func (u *inventoryCRUDUseCasesImpl) DeleteInventory(ctx context.Context, id entities.ID) error {
	if err := u.inventoryRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("unable to delete inventory: %w", err)
	}

	return nil
}

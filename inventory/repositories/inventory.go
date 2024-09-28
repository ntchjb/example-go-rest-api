package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-http-serve/inventory/entities"
	"go-http-serve/utilities/counter"
	"log/slog"
	"time"
)

type InventoryRepository interface {
	Create(ctx context.Context, inventory entities.InventoryData) (entities.ID, error)
	Update(ctx context.Context, id entities.ID, inventoryData entities.PartialInventoryData) error
	Find(ctx context.Context, filter entities.GetInventoriesFilter) (entities.PaginatedInventories, error)
	FindByID(ctx context.Context, id entities.ID) (entities.Inventory, error)
	Delete(ctx context.Context, id entities.ID) error
}

type inventoryRepositoryImpl struct {
	db  *sql.DB
	log *slog.Logger
}

func NewInventoryRepository(db *sql.DB, log *slog.Logger) InventoryRepository {
	return &inventoryRepositoryImpl{
		db:  db,
		log: log,
	}
}

func (r *inventoryRepositoryImpl) Create(ctx context.Context, inventory entities.InventoryData) (entities.ID, error) {
	var resID entities.ID
	now := time.Now().UTC()
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO inventories (name, description, full_price_thb, manufacturer_id, count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;`,
		inventory.Name,
		inventory.Description,
		inventory.FullPriceTHB,
		inventory.ManufacturerID,
		inventory.Count,
		now,
		now,
	).Scan(&resID); err != nil {
		return resID, fmt.Errorf("%w, unable to insert data to database: %v", entities.ErrInternalServer, err)
	}

	return resID, nil
}

func (r *inventoryRepositoryImpl) buildUpdateStatement(id entities.ID, inventoryData entities.PartialInventoryData) (string, []any) {
	res := "UPDATE inventories"
	var numGen counter.Counter
	var setColumns []string
	var params []any
	if inventoryData.Count != nil {
		setColumns = append(setColumns, "count = $"+numGen.NextString())
		params = append(params, *inventoryData.Count)
	}
	if inventoryData.Description != nil {
		setColumns = append(setColumns, "description = $"+numGen.NextString())
		params = append(params, *inventoryData.Description)
	}
	if inventoryData.FullPriceTHB != nil {
		setColumns = append(setColumns, "full_price_thb = $"+numGen.NextString())
		params = append(params, *inventoryData.FullPriceTHB)
	}
	if inventoryData.ManufacturerID != nil {
		setColumns = append(setColumns, "manufacturer_id = $"+numGen.NextString())
		params = append(params, *inventoryData.ManufacturerID)
	}
	if inventoryData.Name != nil {
		setColumns = append(setColumns, "name = $"+numGen.NextString())
		params = append(params, *inventoryData.Name)
	}

	if len(setColumns) > 0 {
		res += " SET"

		now := time.Now().UTC()
		setColumns = append(setColumns, "updated_at = $"+numGen.NextString())
		params = append(params, now)
	}

	for i, col := range setColumns {
		if i > 0 {
			res += ","
		}
		res += " " + col
	}

	res += " WHERE id = $" + numGen.NextString()
	params = append(params, id)

	return res, params
}

func (r *inventoryRepositoryImpl) Update(ctx context.Context, id entities.ID, inventoryData entities.PartialInventoryData) error {
	statement, params := r.buildUpdateStatement(id, inventoryData)
	r.log.Debug("Update SQL", "statement", statement, "params", params)
	result, err := r.db.ExecContext(ctx, statement, params...)
	if err != nil {
		return fmt.Errorf("%w unable to update row: %v", entities.ErrInternalServer, err)
	}
	if rowAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("%w unable to get row affected: %v", entities.ErrInternalServer, err)
	} else if rowAffected == 0 {
		return entities.ErrInventoryNotFound
	}

	return nil
}

func (r *inventoryRepositoryImpl) buildQueryStatement(filter entities.GetInventoriesFilter) (string, []any) {
	var numGen counter.Counter
	res := `SELECT id, name, description, full_price_thb, manufacturer_id, count, created_at, updated_at FROM inventories`
	var params []any
	var whereCondition []string
	if filter.Cursor > 0 {
		whereCondition = append(whereCondition, "id > $"+numGen.NextString())
		params = append(params, filter.Cursor)
	}

	if len(whereCondition) > 0 {
		res += " WHERE"
	}
	for i, cond := range whereCondition {
		if i > 0 {
			res += " AND"
		}
		res += " " + cond
	}

	res += " ORDER BY id"

	if filter.Limit > 0 {
		res += " LIMIT $" + numGen.NextString()
		params = append(params, filter.Limit)
	}

	return res, params
}

func (r *inventoryRepositoryImpl) Find(ctx context.Context, filter entities.GetInventoriesFilter) (entities.PaginatedInventories, error) {
	res := entities.PaginatedInventories{
		Inventories: make([]entities.Inventory, 0),
		Cursor:      0,
	}
	statement, params := r.buildQueryStatement(filter)
	r.log.Debug("Find SQL", "statement", statement, "params", params)

	rows, err := r.db.QueryContext(ctx, statement, params...)
	if err != nil {
		return res, fmt.Errorf("unable to query inventories: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var inventory entities.Inventory
		if err := rows.Scan(&inventory.ID,
			&inventory.Name,
			&inventory.Description,
			&inventory.FullPriceTHB,
			&inventory.ManufacturerID,
			&inventory.Count,
			&inventory.CreatedAt,
			&inventory.UpdatedAt,
		); err != nil {
			return res, fmt.Errorf("%w: unable to scan a row: %v", entities.ErrInternalServer, err)
		}
		res.Inventories = append(res.Inventories, inventory)
	}
	if err := rows.Err(); err != nil {
		return res, fmt.Errorf("%w: error occurred during scanning rows: %v", entities.ErrInternalServer, err)
	}

	if len(res.Inventories) > 0 {
		res.Cursor = res.Inventories[len(res.Inventories)-1].ID
	}

	return res, nil
}

func (r *inventoryRepositoryImpl) FindByID(ctx context.Context, id entities.ID) (entities.Inventory, error) {
	var inventory entities.Inventory
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, full_price_thb, manufacturer_id, count, created_at, updated_at
		FROM inventories
		WHERE id = $1;
	`, id).Scan(
		&inventory.ID,
		&inventory.Name,
		&inventory.Description,
		&inventory.FullPriceTHB,
		&inventory.ManufacturerID,
		&inventory.Count,
		&inventory.CreatedAt,
		&inventory.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return inventory, entities.ErrInventoryNotFound
		}
		return inventory, fmt.Errorf("%w: unable to return inventory: %v", entities.ErrInternalServer, err)
	}

	return inventory, nil
}

func (r *inventoryRepositoryImpl) Delete(ctx context.Context, id entities.ID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM inventories
		WHERE id = $1;
	`,
		id,
	)
	if err != nil {
		return fmt.Errorf("%w unable to delete inventory with ID %v: %v", entities.ErrInternalServer, id, err)
	}

	return nil
}

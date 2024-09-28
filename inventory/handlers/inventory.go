package handlers

import (
	"errors"
	"fmt"
	"go-http-serve/inventory/entities"
	"go-http-serve/inventory/models"
	"go-http-serve/inventory/usecases"
	"go-http-serve/utilities/rest"
	"log/slog"
	"net/http"
	"strconv"
)

type InventoryCRUDHandlers interface {
	CreateInventory(w http.ResponseWriter, r *http.Request)
	DeleteInventory(w http.ResponseWriter, r *http.Request)
	GetInventories(w http.ResponseWriter, r *http.Request)
	UpdateInventory(w http.ResponseWriter, r *http.Request)
}

type inventoryCRUDHandlersImpl struct {
	inventoryCRUDUseCases usecases.InventoryCRUDUseCases
	log                   *slog.Logger
}

func NewInventoryCRUDHandlers(inventoryCRUDUseCases usecases.InventoryCRUDUseCases, log *slog.Logger) InventoryCRUDHandlers {
	return &inventoryCRUDHandlersImpl{
		inventoryCRUDUseCases: inventoryCRUDUseCases,
		log:                   log,
	}
}

func RegisterRoutes(h InventoryCRUDHandlers, mux *http.ServeMux) {
	mux.HandleFunc("POST /inventories", h.CreateInventory)
	mux.HandleFunc("DELETE /inventories/{id}", h.DeleteInventory)
	mux.HandleFunc("GET /inventories", h.GetInventories)
	mux.HandleFunc("PATCH /inventories/{id}", h.UpdateInventory)
}

func (h *inventoryCRUDHandlersImpl) createErrorResponse(err error) (int, rest.ErrorResponse) {
	if errors.Is(err, entities.ErrInventoryNotFound) {
		return http.StatusNotFound, rest.ErrorResponse{
			Error: err.Error(),
		}
	} else {
		h.log.Error("Internal Server Error", "err", err)
		return http.StatusInternalServerError, rest.ErrorResponse{
			Error: "internal server error",
		}
	}
}

func (h *inventoryCRUDHandlersImpl) CreateInventory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var inventoryData entities.InventoryData
	if err := rest.UnmarshalJSON(r, &inventoryData); err != nil {
		resp := rest.ErrorResponse{
			Error: fmt.Sprintf("invalid JSON request: %v", err),
		}
		rest.ReturnJSON(w, http.StatusBadRequest, resp)
		return
	}

	id, err := h.inventoryCRUDUseCases.CreateInventory(ctx, inventoryData)
	if err != nil {
		status, resp := h.createErrorResponse(err)
		rest.ReturnJSON(w, status, resp)
		return
	}

	resp := models.CreateInventoryResponse{
		ID: uint64(id),
	}
	rest.ReturnJSON(w, http.StatusOK, resp)
}

func (h *inventoryCRUDHandlersImpl) DeleteInventory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp := rest.ErrorResponse{
			Error: fmt.Sprintf("invalid inventory ID: %v", idStr),
		}
		rest.ReturnJSON(w, http.StatusBadRequest, resp)
		return
	}

	if err := h.inventoryCRUDUseCases.DeleteInventory(ctx, entities.ID(id)); err != nil {
		status, resp := h.createErrorResponse(err)
		rest.ReturnJSON(w, status, resp)
		return
	}

	rest.Return(w, http.StatusNoContent)
}

func (h *inventoryCRUDHandlersImpl) getQueryForGetInventories(r *http.Request) (entities.GetInventoriesFilter, error) {
	var filter entities.GetInventoriesFilter
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr != "" {
		cursor, err := strconv.ParseUint(cursorStr, 10, 64)
		if err != nil {
			return filter, fmt.Errorf("unable to parse cursor: %w", err)
		}
		filter.Cursor = entities.ID(cursor)
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.ParseUint(limitStr, 10, 64)
		if err != nil {
			return filter, fmt.Errorf("unable to parse limit: %w", err)
		}
		filter.Limit = limit
	}

	return filter, nil
}

func (h *inventoryCRUDHandlersImpl) GetInventories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter, err := h.getQueryForGetInventories(r)
	if err != nil {
		resp := rest.ErrorResponse{
			Error: err.Error(),
		}
		rest.ReturnJSON(w, http.StatusBadRequest, resp)
		return
	}
	inventories, err := h.inventoryCRUDUseCases.GetInventories(ctx, filter)
	if err != nil {
		status, resp := h.createErrorResponse(err)
		rest.ReturnJSON(w, status, resp)
		return
	}

	rest.ReturnJSON(w, http.StatusOK, inventories)
}

func (h *inventoryCRUDHandlersImpl) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp := rest.ErrorResponse{
			Error: fmt.Sprintf("invalid inventory ID: %v", idStr),
		}
		rest.ReturnJSON(w, http.StatusBadRequest, resp)
		return
	}
	var partialInventoryData entities.PartialInventoryData
	if err := rest.UnmarshalJSON(r, &partialInventoryData); err != nil {
		resp := rest.ErrorResponse{
			Error: fmt.Sprintf("invalid JSON request: %v", err),
		}
		rest.ReturnJSON(w, http.StatusBadRequest, resp)
		return
	}

	if err := h.inventoryCRUDUseCases.UpdateInventory(ctx, entities.ID(id), partialInventoryData); err != nil {
		status, resp := h.createErrorResponse(err)
		rest.ReturnJSON(w, status, resp)
		return
	}

	rest.Return(w, http.StatusNoContent)
}

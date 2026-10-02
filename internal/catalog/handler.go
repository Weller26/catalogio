package catalog

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Weller26/catalogio/internal/appcontext"
	"github.com/Weller26/catalogio/internal/httpresponse"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, 
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	
	var req CreateItemRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, "invalid request body",
		)
		return
	}

	item, err := h.service.CreateItem(
		r.Context(),
		userID,
		req,
	)

	if err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, err.Error(),
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	items, err := h.service.ListItems(
		r.Context(),
		userID,
	)

	if err != nil {
		httpresponse.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, http.StatusUnauthorized, "unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	itemID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, "invalid item id",
		)
		return
	}

	item, err := h.service.GetItemByID(
		r.Context(),
		itemID,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpresponse.WriteError(
				w, http.StatusNotFound, "item not found",
			)
			return
		}

		httpresponse.WriteError(
			w, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) UpdateItemByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, http.StatusUnauthorized, "unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	itemID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, "invalid item id",
		)
		return
	}

	var req UpdateItemRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, "invalid request body",
		)
		return
	}

	item, err := h.service.UpdateItemByID(
		r.Context(),
		itemID,
		userID,
		req,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpresponse.WriteError(
				w, http.StatusNotFound, "item not found",
			)
			return
		}

		httpresponse.WriteError(
			w, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteItemByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, http.StatusUnauthorized, "unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	itemID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, "invalid item id",
		)
		return
	}

	item, err := h.service.DeleteItemByID(
		r.Context(),
		itemID,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpresponse.WriteError(
				w, http.StatusNotFound, "item not found",
			)
			return
		}

		httpresponse.WriteError(
			w, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateItemStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, 
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	
	if err := decoder.Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	itemStatus, err := h.service.CreateItemStatus(
		r.Context(),
		userID,
		req.Name,
	)

	if err != nil {
		httpresponse.WriteError(
			w, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, itemStatus)
}

func (h *Handler) ListItemStatuses(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	itemStatuses, err := h.service.ListItemStatuses(
		r.Context(),
		userID,
	)

	if err != nil {
		httpresponse.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, itemStatuses)
}

func (h *Handler) DeleteItemStatusByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, http.StatusUnauthorized, "unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	itemStatusID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, "invalid item status id",
		)
		return
	}

	itemStatus, err := h.service.DeleteItemStatusByID(
		r.Context(),
		itemStatusID,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpresponse.WriteError(
				w, http.StatusNotFound, "item status not found",
			)
			return
		}

		httpresponse.WriteError(
			w, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, itemStatus)
}

func (h *Handler) CreateItemType(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, 
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	
	if err := decoder.Decode(&req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	itemType, err := h.service.CreateItemType(
		r.Context(),
		userID,
		req.Name,
	)

	if err != nil {
		httpresponse.WriteError(
			w, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, itemType)
}

func (h *Handler) ListItemTypes(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	itemTypes, err := h.service.ListItemTypes(
		r.Context(),
		userID,
	)

	if err != nil {
		httpresponse.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, itemTypes)
}

func (h *Handler) DeleteItemTypeByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := appcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(
			w, http.StatusUnauthorized, "unauthorized",
		)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	itemTypeID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpresponse.WriteError(
			w, http.StatusBadRequest, "invalid item type id",
		)
		return
	}

	itemType, err := h.service.DeleteItemTypeByID(
		r.Context(),
		itemTypeID,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpresponse.WriteError(
				w, http.StatusNotFound, "item type not found",
			)
			return
		}

		httpresponse.WriteError(
			w, http.StatusInternalServerError, "internal server error",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, itemType)
}
package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type DataResponse struct {
	Data interface{} `json:"data"`
}

type ListResponse struct {
	Data       interface{}     `json:"data"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type ErrorResponse struct {
	Error *AppError `json:"error"`
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(DataResponse{Data: data}); err != nil {
			slog.Error("failed to encode response", "error", err)
		}
	}
}

func RespondList(w http.ResponseWriter, data interface{}, pagination *PaginationMeta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ListResponse{Data: data, Pagination: pagination}); err != nil {
		slog.Error("failed to encode list response", "error", err)
	}
}

func RespondError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*AppError)
	if !ok {
		appErr = ErrInternal("an unexpected error occurred")
		slog.Error("unhandled error", "error", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: appErr})
}

func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// DecodeJSON decodes a JSON request body into the given struct.
func DecodeJSON(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return ErrValidation("invalid request body", nil)
	}
	return nil
}

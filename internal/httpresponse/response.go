package httpresponse

import (
	"context"
	"encoding/json"
	"net/http"
	"log/slog"
)

type errorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	clientMessage string,
	err error,
) {
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(ctx, "internal error",
			"status", status,
			"error", err,
			"client_message", clientMessage,
		)
	} else if status == http.StatusUnauthorized && err == nil {
		slog.DebugContext(ctx, "client error",
			"status", status,
			"client_message", clientMessage,
		)
	} else {
		slog.WarnContext(ctx, "client error",
			"status", status,
			"error", err,
			"client_message", clientMessage,
		)
	}

	WriteJSON(
		w,
		status,
		errorResponse{
			Error: clientMessage,
		},
	)
}
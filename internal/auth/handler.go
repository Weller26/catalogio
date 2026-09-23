package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Handler struct {
	service *Service
	secureCookies bool
}

func NewHandler(
	service *Service,
	secureCookies bool,
) *Handler {
	return &Handler{
		service: service,
		secureCookies: secureCookies,
	}
}

func (h *Handler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req RegisterRequest

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1<<20,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid JSON",
		)
		return
	}

	user, err := h.service.Register(
		r.Context(),
		req,
	)

	if errors.Is(err, ErrEmailTaken) {
		writeError(
			w,
			http.StatusConflict,
			"email already exists",
		)
		return
	}

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		user,
	)
}

func (h *Handler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req LoginRequest

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1<<20,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid JSON",
		)
		return
	}

	user, token, err := h.service.Login(
		r.Context(),
		req,
	)

	if errors.Is(err, ErrInvalidCredentials) {
		writeError(
			w,
			http.StatusUnauthorized,
			"invalid email or password",
		)
		return
	}

	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	h.setSessionCookie(w, token)

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	cookie, err := r.Cookie(
		"catalogio_session",
	)

	if err == nil {
		_ = h.service.Logout(
			r.Context(),
			cookie.Value,
		)
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name: "catalogio_session",
			Value: "",
			Path: "/",
			HttpOnly: true,
			Secure: h.secureCookies,
			SameSite: http.SameSiteLaxMode,
			MaxAge: -1,
		},
	)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, ok := UserFromContext(
		r.Context(),
	)

	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		user,
	)
}

func (h *Handler) setSessionCookie(
	w http.ResponseWriter,
	token string,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name: "catalogio_session",
			Value: token,
			Path: "/",
			HttpOnly: true,
			Secure: h.secureCookies,
			SameSite: http.SameSiteLaxMode,
			MaxAge: int((7 * 24 * time.Hour).Seconds()),
		},
	)
}
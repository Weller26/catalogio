package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Weller26/catalogio/internal/appcontext"
	"github.com/Weller26/catalogio/internal/httpresponse"
	"github.com/Weller26/catalogio/internal/metrics"
)

const refreshTokenCookie = "catalogio_refresh_token"

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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httpresponse.WriteError(
			r.Context(), w, http.StatusBadRequest,
			"registration failed: invalid JSON", err,
		)
		return
	}

	user, err := h.service.Register(
		r.Context(),
		req,
	)

	if errors.Is(err, ErrEmailTaken) {
		httpresponse.WriteError(
			r.Context(), w, http.StatusConflict,
			"email already exists", err,
		)
		return
	}

	if err != nil {
		httpresponse.WriteError(
			r.Context(), w, http.StatusBadRequest,
			err.Error(), err,
		)
		return
	}

	metrics.UsersRegisteredTotal.Inc()

	slog.InfoContext(r.Context(), "user registered",
		"user_id", user.ID,
	)

	httpresponse.WriteJSON(
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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httpresponse.WriteError(
			r.Context(), w, http.StatusBadRequest,
			"invalid JSON", err,
		)
		return
	}

	user, accessToken, refreshToken, err := h.service.Login(
		r.Context(),
		req,
	)

	if errors.Is(err, ErrInvalidCredentials) {
		metrics.AuthAttemptsTotal.WithLabelValues("failure").Inc()
		httpresponse.WriteError(
			r.Context(), w, http.StatusUnauthorized,
			"login failed: invalid credentials", err,
		)
		return
	}

	if err != nil {
		metrics.AuthAttemptsTotal.WithLabelValues("failure").Inc()
		httpresponse.WriteError(
			r.Context(), w, http.StatusInternalServerError,
			"internal server error", err,
		)
		return
	}

	h.setRefreshTokenCookie(w, refreshToken)

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"user": user,
		"access_token": accessToken,
	})

	metrics.AuthAttemptsTotal.WithLabelValues("success").Inc()

	slog.InfoContext(r.Context(), "user logged in",
		"user_id", user.ID.String(),
	)
}

func (h *Handler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	cookie, err := r.Cookie(refreshTokenCookie)

	if err == nil {
		err = h.service.Logout(
			r.Context(),
			cookie.Value,
		)

		if err != nil {
			httpresponse.WriteError(
				r.Context(), w, http.StatusInternalServerError,
				"internal server error", err,
			)
			return
		}
	}

	h.deleteCookie(w, refreshTokenCookie)

	httpresponse.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "logged out",
	})
}

func (h *Handler) Refresh(
	w http.ResponseWriter,
	r *http.Request,
) {
	cookie, err := r.Cookie(refreshTokenCookie)

	if err != nil {
		httpresponse.WriteError(
			r.Context(), w, http.StatusUnauthorized,
			"refresh token required", err,
		)
		return
	}

	accessToken, err := h.service.Refresh(
		r.Context(),
		cookie.Value,
	)

	if errors.Is(err, ErrInvalidCredentials) {
		httpresponse.WriteError(
			r.Context(), w, http.StatusUnauthorized,
			"invalid refresh token", err,
		)
		return
	}

	if err != nil {
		httpresponse.WriteError(
			r.Context(), w, http.StatusInternalServerError,
			"internal server error", err,
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, map[string]string{
			"access_token": accessToken,
		},
	)
}

func (h *Handler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := appcontext.UserIDFromContext(
		r.Context(),
	)

	if !ok {
		httpresponse.WriteError(
			r.Context(), w, http.StatusUnauthorized,
			"unauthorized", nil,
		)
		return
	}

	user, err := h.service.GetUserByID(
		r.Context(),
		userID,
	)

	if err != nil {
		httpresponse.WriteError(
			r.Context(), w, http.StatusInternalServerError,
			"internal server error", nil,
		)
		return
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		user,
	)
}

func (h *Handler) setRefreshTokenCookie(
	w http.ResponseWriter,
	token string,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name: refreshTokenCookie,
			Value: token,
			Path: "/",
			HttpOnly: true,
			Secure: h.secureCookies,
			SameSite: http.SameSiteLaxMode,
			MaxAge: int((24 * time.Hour).Seconds()),
		},
	)
}

func (h *Handler) deleteCookie(
	w http.ResponseWriter,
	name string,
) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name: name,
			Value: "",
			Path: "/",
			HttpOnly: true,
			Secure: h.secureCookies,
			SameSite: http.SameSiteLaxMode,
			MaxAge: -1,
		},
	)
}
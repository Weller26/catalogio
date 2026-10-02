package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Weller26/catalogio/internal/appcontext"
	"github.com/Weller26/catalogio/internal/httpresponse"
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
		httpresponse.WriteError(
			w,
			http.StatusConflict,
			"email already exists",
		)
		return
	}

	if err != nil {
		httpresponse.WriteError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

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
			w,
			http.StatusBadRequest,
			"invalid JSON",
		)
		return
	}

	user, accessToken, refreshToken, err := h.service.Login(
		r.Context(),
		req,
	)

	if errors.Is(err, ErrInvalidCredentials) {
		httpresponse.WriteError(
			w,
			http.StatusUnauthorized,
			"invalid email or password",
		)
		return
	}

	if err != nil {
		httpresponse.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	h.setRefreshTokenCookie(w, refreshToken)

	httpresponse.WriteJSON(w, http.StatusOK, map[string]any{
		"user": user,
		"access_token": accessToken,
	})

	log.Printf("user %s has logged in", req.Email)
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
				w,
				http.StatusInternalServerError,
				"internal server error",
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
			w,
			http.StatusUnauthorized,
			"refresh token required",
		)
		return
	}

	accessToken, err := h.service.Refresh(
		r.Context(),
		cookie.Value,
	)

	if errors.Is(err, ErrInvalidCredentials) {
		httpresponse.WriteError(
			w,
			http.StatusUnauthorized,
			"invalid refresh token",
		)
		return
	}

	if err != nil {
		httpresponse.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
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
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	user, err := h.service.GetUserByID(
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
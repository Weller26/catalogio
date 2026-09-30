package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const (
	// accessTokenCookie = "catalogio_access_token"
	refreshTokenCookie = "catalogio_refresh_token"
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

	user, accessToken, refreshToken, err := h.service.Login(
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

	//h.setAccessTokenCookie(w, accessToken)
	h.setRefreshTokenCookie(w, refreshToken)

	writeJSON(w, http.StatusOK, map[string]any{
		"user": user,
		"access_token": accessToken,
	})
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
			writeError(
				w,
				http.StatusInternalServerError,
				"internal server error",
			)
			return
		}
	}

	h.deleteCookie(w, refreshTokenCookie)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logged out",
	})
}

func (h *Handler) Refresh(
	w http.ResponseWriter,
	r *http.Request,
) {
	cookie, err := r.Cookie(refreshTokenCookie)

	if err != nil {
		writeError(
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
		writeError(
			w,
			http.StatusUnauthorized,
			"invalid refresh token",
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

	writeJSON(w, http.StatusOK, map[string]string{
			"access_token": accessToken,
		},
	)
}

func (h *Handler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := UserIDFromContext(
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

	user, err := h.service.repo.GetUserByID(
		r.Context(),
		userID,
	)

	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		user,
	)
}

// func (h *Handler) setAccessTokenCookie(
// 	w http.ResponseWriter,
// 	token string,
// ) {
// 	http.SetCookie(
// 		w,
// 		&http.Cookie{
// 			Name: accessTokenCookie,
// 			Value: token,
// 			Path: "/",
// 			HttpOnly: true,
// 			Secure: h.secureCookies,
// 			SameSite: http.SameSiteLaxMode,
// 			MaxAge: int((15 * time.Minute).Seconds()),
// 		},
// 	)
// }

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
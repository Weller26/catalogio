package httpapi

import (
	"net/http"

	"github.com/Weller26/catalogio/internal/auth"
	"github.com/Weller26/catalogio/internal/items"
)

func NewRouter(
	authService *auth.Service,
	authHandler *auth.Handler,
	itemHandler *items.Handler,
) http.Handler {
	mux := http.NewServeMux()

	// api

	mux.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/logout",
		authHandler.Logout,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/refresh",
		authHandler.Refresh,
	)

	meHandler := http.HandlerFunc(
		authHandler.Me,
	)

	mux.Handle(
		"GET /api/v1/me",
		authService.RequireAuth(meHandler),
	)

	mux.Handle(
		"POST /api/v1/items",
		authService.RequireAuth(
			http.HandlerFunc(itemHandler.Create),
		),
	)

	mux.Handle(
		"GET /api/v1/items",
		authService.RequireAuth(
			http.HandlerFunc(itemHandler.ListItems),
		),
	)

	mux.Handle(
		"GET /api/v1/items/{id}",
		authService.RequireAuth(
			http.HandlerFunc(itemHandler.GetItemByID),
		),
	)

	mux.Handle(
		"PUT /api/v1/items/{id}",
		authService.RequireAuth(
			http.HandlerFunc(itemHandler.UpdateItemByID),
		),
	)

	mux.Handle(
		"DELETE /api/v1/items/{id}",
		authService.RequireAuth(
			http.HandlerFunc(itemHandler.DeleteItemByID),
		),
	)

	// frontend

	mux.Handle(
        "GET /static/",
        http.StripPrefix(
            "/static/",
            http.FileServer(
                http.Dir("./web/static"),
            ),
        ),
    )

    mux.Handle(
        "GET /login.html",
        http.FileServer(
            http.Dir("./web"),
        ),
    )

    mux.Handle(
        "GET /register.html",
        http.FileServer(
            http.Dir("./web"),
        ),
    )

    mux.Handle(
        "GET /items.html",
        http.FileServer(
            http.Dir("./web"),
        ),
    )

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(
			w,
			r,
			"./web/login.html",
		)
	})

		mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(
			w,
			r,
			"./web/register.html",
		)
	})

		mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(
			w,
			r,
			"./web/items.html",
		)
	})

	return mux
}
package httpapi

import (
	"net/http"

	"github.com/Weller26/catalogio/internal/auth"
	"github.com/Weller26/catalogio/internal/catalog"
)

func NewRouter(
	authHandler *auth.Handler,
	itemHandler *catalog.Handler,
	authMiddleware *AuthMiddleware,
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
		authMiddleware.requireAuth(meHandler),
	)

	mux.Handle(
		"POST /api/v1/items",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.CreateItem),
		),
	)

	mux.Handle(
		"GET /api/v1/items",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.ListItems),
		),
	)

	mux.Handle(
		"GET /api/v1/items/{id}",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.GetItemByID),
		),
	)

	mux.Handle(
		"PUT /api/v1/items/{id}",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.UpdateItemByID),
		),
	)

	mux.Handle(
		"DELETE /api/v1/items/{id}",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.DeleteItemByID),
		),
	)

	mux.Handle(
		"POST /api/v1/item-statuses",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.CreateItemStatus),
		),
	)

	mux.Handle(
		"GET /api/v1/item-statuses",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.ListItemStatuses),
		),
	)

	mux.Handle(
		"DELETE /api/v1/item-statuses/{id}",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.DeleteItemStatusByID),
		),
	)

	mux.Handle(
		"POST /api/v1/item-types",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.CreateItemType),
		),
	)

	mux.Handle(
		"GET /api/v1/item-types",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.ListItemTypes),
		),
	)

	mux.Handle(
		"DELETE /api/v1/item-types/{id}",
		authMiddleware.requireAuth(
			http.HandlerFunc(itemHandler.DeleteItemTypeByID),
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
        "GET /catalog.html",
        http.FileServer(
            http.Dir("./web"),
        ),
    )

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/login.html")
	})

	mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/register.html")
	})

	mux.HandleFunc("GET /catalog", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/catalog.html")
	})

	return mux
}
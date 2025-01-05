package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/monolithic/internal/core/domain"
	"example.com/monolithic/internal/core/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// Routes sets up the user routes
func (h *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createUser)     // POST /api/users
	r.Get("/{userID}", h.getUser) // GET /api/users/{userID}
	return r
}

// CreateUser handles user creation
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}
	defer r.Body.Close()

	err := h.service.CreateUser(r.Context(), &user)
	if err != nil {
		switch err {
		case services.ErrInvalidInput:
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": err.Error()})
		case services.ErrDuplicateEmail:
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, map[string]string{"error": err.Error()})
		default:
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Internal server error"})
		}
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, user)
}

// GetUser handles fetching a single user
func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "User ID is required"})
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, map[string]string{"error": "User not found"})
		default:
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Internal server error"})
		}
		return
	}

	render.JSON(w, r, user)
}

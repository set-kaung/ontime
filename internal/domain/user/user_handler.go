package user

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/set-kaung/senior_project_1/internal/helpers"
	"github.com/set-kaung/senior_project_1/internal/util"
	"modernc.org/sqlite"
)

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")

type UserHandler struct {
	UserService    UserService
	SessionManager *scs.SessionManager
}

func (h *UserHandler) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	type RegisterUser struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	user := RegisterUser{}
	err := decoder.Decode(&user)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid json request", nil)
		return
	}
	em, err := NewEmail(user.Email)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid email address", nil)
		return
	}
	err = h.UserService.InsertUser(r.Context(), em, user.Username, user.Password)
	if err != nil {
		if errors.Is(err, &sqlite.Error{}) {
			log.Println(err)
			helpers.WriteError(w, http.StatusInternalServerError, "Sorry. Something Happend on the server.", nil)
			return
		}
		helpers.WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
}

func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	user := LoginUser{}
	err := decoder.Decode(&user)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "Invalid JSON request", nil)
		return
	}
	id, result := h.UserService.AuthenticateWithDB(r.Context(), user.Email, user.Password)
	if !result {
		helpers.WriteError(w, http.StatusNotFound, "User Not Found", nil)
		return
	}
	err = h.SessionManager.RenewToken(r.Context())
	if err != nil {
		log.Println(err)
		helpers.WriteError(w, http.StatusInternalServerError, "Server is having problems", nil)
		return
	}
	h.SessionManager.Put(r.Context(), "authenticatedUserID", int(id))
	helpers.WriteSuccess(w, http.StatusOK, "", nil)
}

func (h *UserHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	err := h.SessionManager.RenewToken(r.Context())
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "server is having problems", nil)
		return
	}
	h.SessionManager.Remove(r.Context(), "authenticatedUserID")
}

func (h *UserHandler) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := h.SessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}
		exists, err := h.UserService.Exists(r.Context(), int32(id))
		if err != nil {
			helpers.WriteError(w, http.StatusInternalServerError, "Server is having problems", nil)
			return
		}
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func isAuthenticated(r *http.Request) bool {
	isAuthenticatedBoolean, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticatedBoolean
}

func (h *UserHandler) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			helpers.WriteError(w, http.StatusForbidden, "You need to login", nil)
			return
		}

		//Set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})

}

func (us *UserService) AuthenticateWithDB(ctx context.Context, email string, password string) (int32, bool) {
	user, err := us.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Println(err)
		return -1, false
	}
	return user.ID, util.CheckPasswordHash(password, user.Password)
}

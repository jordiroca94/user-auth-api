package user

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jordiroca94/moviechase-api/config"
	"github.com/jordiroca94/moviechase-api/service/auth"
	"github.com/jordiroca94/moviechase-api/types"
	"github.com/jordiroca94/moviechase-api/utils"
)

type UserHandler struct {
	service    *UserService
	repository types.UserStore
}

func NewHandler(service *UserService, repository types.UserStore) *UserHandler {
	return &UserHandler{
		service:    service,
		repository: repository}
}

func (h *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserPayload
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %s", errors))
		return
	}

	u, err := h.service.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found, invalid email or password"))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid email or password"))
		return
	}

	secret := []byte(config.Envs.JWTSecret)
	token, err := h.service.CreateToken(secret, u.ID, u.Email, u.FirstName, u.LastName)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})

}

func (h *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {

	var payload types.RegisterUserPayload
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %s", errors))
		return
	}

	_, err := h.service.GetUserByEmail(payload.Email)
	if err == nil {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.service.CreateUser(payload, hashedPassword)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": "User created successfully"})
}

func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var payload types.UpdateUserPayload
	if err := utils.ParseJson(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %s", errors))
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user id"))
	}

	user, err := h.service.GetUserByID(idInt)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}

	if user.Email != payload.Email {
		_, err := h.service.GetUserByEmail(payload.Email)
		if err == nil {
			utils.WriteError(w, http.StatusConflict, fmt.Errorf("user with email %s already exists", payload.Email))
			return
		}
	}

	err = h.service.UpdateUser(idInt, payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

func (h *UserHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user id"))
	}

	user, err := h.service.GetUserByID(idInt)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("user not found"))
		return
	}
	utils.WriteJSON(w, http.StatusOK, user)
}
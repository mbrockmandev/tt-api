package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	auth "github.com/mbrockmandev/tometracker/internal/auth"
	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
	"github.com/mbrockmandev/tometracker/internal/models"
)

func (h *Handler) RegisterNewUser(w http.ResponseWriter,
	r *http.Request) {

	var reqPayload struct {
		Email           string `json:"email"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		Role            string `json:"role"`
	}

	err := jsonHelper.ReadJson(w, r, &reqPayload)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if reqPayload.Password != reqPayload.ConfirmPassword {
		jsonHelper.ErrorJson(w, errors.New("passwords do not match"), http.StatusBadRequest)
		return
	}

	existingUser, _ := h.App.DB.GetUserByEmail(strings.ToLower(reqPayload.Email))
	if existingUser != nil {
		jsonHelper.ErrorJson(w, errors.New("email already registered"), http.StatusConflict)
		return
	}

	hashedPw, err := h.App.Auth.HashPassword(strings.ToLower(reqPayload.Password))
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		Email:     reqPayload.Email,
		FirstName: reqPayload.FirstName,
		LastName:  reqPayload.LastName,
		Password:  hashedPw,
		Role:      reqPayload.Role,
	}

	id, err := h.App.DB.CreateUser(&newUser)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	if id == 0 {
		jsonHelper.ErrorJson(w, errors.New("failed to create user"), http.StatusInternalServerError)
		return
	}

	u := auth.JwtUser{
		ID:    id,
		Email: newUser.Email,
		Role:  newUser.Role,
	}

	tokens, accessCookie, err := h.App.Auth.GenerateTokenPairAndCookie(&u)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	refreshCookie := h.App.Auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, accessCookie)

	response := struct {
		AccessToken string       `json:"access_token"`
		UserInfo    auth.JwtUser `json:"user_info"`
	}{
		AccessToken: tokens.AccessToken,
		UserInfo:    u,
	}

	jsonHelper.WriteJson(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var reqPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonHelper.ReadJson(w, r, &reqPayload)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	user, err := h.App.DB.GetUserByEmail(reqPayload.Email)
	if err != nil {
		jsonHelper.ErrorJson(w, errors.New("invalid credentials (email)"), http.StatusBadRequest)
		return
	}

	valid, err := user.ValidPassword(reqPayload.Password)
	if err != nil || !valid {
		jsonHelper.ErrorJson(w, errors.New("invalid credentials (password)"), http.StatusBadRequest)
		return
	}

	u := auth.JwtUser{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	tokens, accessCookie, err := h.App.Auth.GenerateTokenPairAndCookie(&u)
	if err != nil {
		jsonHelper.ErrorJson(w, err)
		return
	}

	refreshCookie := h.App.Auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, accessCookie)

	response := struct {
		AccessToken string       `json:"access_token"`
		UserInfo    auth.JwtUser `json:"user_info"`
	}{
		AccessToken: tokens.AccessToken,
		UserInfo:    u,
	}

	jsonHelper.WriteJson(w, http.StatusAccepted, response)
}

func (h *Handler) GetRefreshToken(w http.ResponseWriter, r *http.Request) {

	_, claims, err := h.App.Auth.GetTokenAndVerify(w, r)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("unauthorized, unable to verify token validity: %s", err), http.StatusUnauthorized)
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("userId unable to be extracted from token: %s", err), http.StatusUnauthorized)
		return
	}

	user, err := h.App.DB.GetUserById(userId)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("unable to find user with ID: %s", err), http.StatusUnauthorized)
		return
	}

	u := auth.JwtUser{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	tokens, accessCookie, err := h.App.Auth.GenerateTokenPairAndCookie(&u)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("error generating tokens: %s", err), http.StatusUnauthorized)
		return
	}

	refreshCookie := h.App.Auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, accessCookie)

	response := struct {
		AccessToken string       `json:"access_token"`
		UserInfo    auth.JwtUser `json:"user_info"`
	}{
		AccessToken: tokens.AccessToken,
		UserInfo:    u,
	}

	jsonHelper.WriteJson(w, http.StatusAccepted, response)

}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.App.Auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetUserInfoFromJWT(w http.ResponseWriter, r *http.Request) {
	_, claims, err := h.App.Auth.GetTokenAndVerify(w, r)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("unauthorized: %s", err), http.StatusUnauthorized)
		return
	}

	userId, err := strconv.Atoi(claims.Subject)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("unknown user: %s", err), http.StatusUnauthorized)
		return
	}

	user, err := h.App.DB.GetUserById(userId)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("unknown user: %s", err), http.StatusUnauthorized)
		return
	}

	u := auth.JwtUser{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	tokens, accessCookie, err := h.App.Auth.GenerateTokenPairAndCookie(&u)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("error generating tokens: %s", err), http.StatusUnauthorized)
		return
	}

	refreshCookie := h.App.Auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)
	http.SetCookie(w, accessCookie)

	response := struct {
		AccessToken string       `json:"access_token"`
		UserInfo    auth.JwtUser `json:"user_info"`
	}{
		AccessToken: tokens.AccessToken,
		UserInfo:    u,
	}

	jsonHelper.WriteJson(w, http.StatusAccepted, response)
}

// func (h *Handler) getUserIdFromJWT(r *http.Request) (int, error) {
// 	_, claims, err := h.App.Auth.GetTokenAndVerify(nil, r)
// 	if err != nil {
// 		return 0, err
// 	}

// 	userId, err := strconv.Atoi(claims.Subject)
// 	if err != nil {
// 		return 0, fmt.Errorf("invalid user id in token")
// 	}

// 	return userId, nil
// }

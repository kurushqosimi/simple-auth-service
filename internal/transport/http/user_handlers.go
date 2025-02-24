package http

import (
	"errors"
	"fullstack-simple-app/internal/errcode"
	"fullstack-simple-app/internal/models"
	"fullstack-simple-app/pkg/app_errors"
	"fullstack-simple-app/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	userService UserService
	logger      logger.Logger
}

type UserService interface {
	RegisterUser(user *models.User, password string) error
	VerifyUser(email string, otp string) (models.User, string, error)
	ResendCode(email string) error
	UserSignIn(email string, password string) (string, error)
	GetUser(email string) (models.User, error)
}

func NewUserHandler(userService UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

type registerUserRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" biding:"required"`
}

func (h *UserHandler) RegisterUserHandler(ctx *gin.Context) {
	const op = "registerUserHandler"

	var req registerUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("%s: ShouldBindJSON: %v", op, err)
		respondWithError(ctx, http.StatusBadRequest, errcode.ErrInvalidRequest, "", nil)
		return
	}

	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	err := h.userService.RegisterUser(user, req.Password)
	if err != nil {
		h.logger.Error("%s: h.userService.RegisterUser: %v", op, err)

		var appErr *app_errors.AppError
		if errors.As(err, &appErr) {
			statusCode := statusFromCode(appErr.Code)
			respondWithError(ctx, statusCode, appErr.Code, "", appErr)
		} else {
			respondWithError(ctx, http.StatusInternalServerError, errcode.ErrInternal, "", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

type verifyUserRequest struct {
	Code  string `json:"code" binding:"required"`
	Email string `json:"email" binding:"required"`
}

const loginHTMLPath = "../html/login.html"

func (h *UserHandler) VerifyUserHandler(ctx *gin.Context) {
	const op = "VerifyUserHandler"

	var req verifyUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("%s: ShouldBindJSON: %v", op, err)
		respondWithError(ctx, http.StatusBadRequest, errcode.ErrInvalidRequest, "", nil)
		return
	}

	user, token, err := h.userService.VerifyUser(req.Email, req.Code)
	if err != nil {
		h.logger.Error("%s: h.userService.VerifyUser: %v", op, err)

		var appErr *app_errors.AppError
		if errors.As(err, &appErr) {
			statusCode := statusFromCode(appErr.Code)
			if statusCode == http.StatusFound {
				ctx.Redirect(http.StatusFound, loginHTMLPath)
				return
			}
			respondWithError(ctx, statusCode, appErr.Code, "", appErr)
		} else {
			respondWithError(ctx, http.StatusInternalServerError, errcode.ErrInternal, "", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user, "access-token": token})
}

type resendCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *UserHandler) ResendCodeHandler(ctx *gin.Context) {
	const op = "ResendCodeHandler"

	var req resendCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("%s: ShouldBindJSON: %v", op, err)
		respondWithError(ctx, http.StatusBadRequest, errcode.ErrInvalidRequest, "", nil)
		return
	}

	err := h.userService.ResendCode(req.Email)
	if err != nil {
		h.logger.Error("%s: h.userService.ResendCode: %v", op, err)

		var appErr *app_errors.AppError
		if errors.As(err, &appErr) {
			statusCode := statusFromCode(appErr.Code)
			respondWithError(ctx, statusCode, appErr.Code, "", appErr)
		} else {
			respondWithError(ctx, http.StatusInternalServerError, errcode.ErrInternal, "", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "verification was sent"})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandler) LoginHandler(ctx *gin.Context) {
	const op = "LoginHandler"

	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("%s: ShouldBindJSON: %v", op, err)
		respondWithError(ctx, http.StatusBadRequest, errcode.ErrInvalidRequest, "", nil)
		return
	}

	accessToken, err := h.userService.UserSignIn(req.Email, req.Password)
	if err != nil {
		h.logger.Error("%s: h.userService.UserSignIn: %v", op, err)

		var appErr *app_errors.AppError
		if errors.As(err, &appErr) {
			statusCode := statusFromCode(appErr.Code)
			respondWithError(ctx, statusCode, appErr.Code, "", appErr)
		} else {
			respondWithError(ctx, http.StatusInternalServerError, errcode.ErrInternal, "", err)
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"access-token": accessToken})
}

type getUserRequest struct {
	Email string `uri:"email" binding:"required"`
}

func (h *UserHandler) GetUserHandler(ctx *gin.Context) {
	const op = "GetUserHandler"

	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		h.logger.Error("%s: ShouldBindJSON: %v", op, err)
		respondWithError(ctx, http.StatusBadRequest, errcode.ErrInvalidRequest, "", nil)
		return
	}

	user, err := h.userService.GetUser(req.Email)
	if err != nil {
		h.logger.Error("%s: h.userService.GetUser: %v", op, err)

		var appErr *app_errors.AppError
		if errors.As(err, &appErr) {
			statusCode := statusFromCode(appErr.Code)
			respondWithError(ctx, statusCode, appErr.Code, "", appErr)
		} else {
			respondWithError(ctx, http.StatusInternalServerError, errcode.ErrInternal, "", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

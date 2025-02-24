package http

import "github.com/gin-gonic/gin"

func NewRouter(userHandler *UserHandler) *gin.Engine {
	r := gin.Default()

	registerUserRoutes(r, userHandler)

	return r
}

func registerUserRoutes(r *gin.Engine, h *UserHandler) {
	r.POST("/users", h.RegisterUserHandler)
	r.PATCH("/users/activate", h.VerifyUserHandler)
	r.PATCH("/users/resend-code", h.ResendCodeHandler)
	r.POST("/users/login", h.LoginHandler)
	r.GET("/users/:email", h.GetUserHandler)
}

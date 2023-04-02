package api

import (
	"net/http"
	db "simplebank/db/sqlc"

	"github.com/gin-gonic/gin"
)

type createUserReq struct {
	Username       string `json:"username" binding:"required"`
	HashedPassword string `json:"hashed_password" binding:"required"`
	FullName       string `json:"full_name" binding:"required"`
	Email          string `json:"email" binding:"required"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: req.HashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	account, err := server.store.CreateUser(ctx, arg)

	if err != nil {

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

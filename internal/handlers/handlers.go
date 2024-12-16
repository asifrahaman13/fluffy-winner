package handlers

import (
	"fmt"
	"github.com/asifrahaman13/bhagabad_gita/internal/core/domain"
	"github.com/asifrahaman13/bhagabad_gita/internal/core/ports"
	"github.com/asifrahaman13/bhagabad_gita/internal/helper"
	"github.com/gin-gonic/gin"
)

var UserHandler *userHandler

type userHandler struct {
	userService ports.UserService
}

func (h *userHandler) Initialize(userserv ports.UserService) {
	UserHandler = &userHandler{
		userService: userserv,
	}
}

func (h *userHandler) Signup(c *gin.Context) {
	var user domain.User
	c.BindJSON(&user)
	message, err := h.userService.Signup(user)
	if err != nil {
		panic(err)
	}
	helper.JSONResponse(c, 200, message, nil)
}

func (h *userHandler) Login(c *gin.Context) {
	var user domain.User
	c.BindJSON(&user)
	message, err := h.userService.Login(user)
	if err != nil {
		panic(err)
	}
	helper.JSONResponse(c, 200, message, nil)
}

func (s *userHandler) PublicApi(c *gin.Context) {
	message := make(map[string]string)
	var userSearch domain.Query
	c.BindJSON(&userSearch)
	fmt.Println(userSearch.Search)
	llmResponse, err := s.userService.GetLLMResponse(userSearch.Search)
	if err != nil {
		panic(err)
	}
	message["message"] = llmResponse
	helper.JSONResponse(c, 200, message, nil)
}

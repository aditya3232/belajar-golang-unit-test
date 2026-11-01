package handler

import (
	"errors"

	"belajar-golang-unit-test/config"
	"belajar-golang-unit-test/internal/adapter"
	"belajar-golang-unit-test/internal/adapter/handler/request"
	"belajar-golang-unit-test/internal/adapter/handler/response"
	"belajar-golang-unit-test/internal/adapter/repository"
	"belajar-golang-unit-test/internal/core/domain/entity"
	"belajar-golang-unit-test/internal/core/service"
	"belajar-golang-unit-test/utils/conv"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type UserHandlerInterface interface {
	SignUp(c fiber.Ctx) error
	SignIn(c fiber.Ctx) error
	GetProfile(c fiber.Ctx) error
	UpdateProfile(c fiber.Ctx) error
	DeleteUser(c fiber.Ctx) error
}

type userHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(
	app *fiber.App,
	userService service.UserServiceInterface,
	cfg *config.Config,
	jwtService service.JwtServiceInterface,
	redisClient *redis.Client,
) UserHandlerInterface {
	h := &userHandler{userService: userService}
	mid := adapter.NewMiddlewareAdapter(cfg, jwtService, redisClient)

	// Public routes
	api := app.Group("/api/v1")
	api.Post("/auth/signup", h.SignUp)
	api.Post("/auth/signin", h.SignIn)

	// Protected routes (JWT + Redis session)
	auth := api.Group("/users", mid.CheckToken())
	auth.Get("/profile", h.GetProfile)
	auth.Put("/profile", h.UpdateProfile)
	auth.Delete("/:id", h.DeleteUser)

	return h
}

func (h *userHandler) SignUp(c fiber.Ctx) error {
	var req request.SignUpRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Warn().Err(err).Msg("failed to bind signup request")
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	userEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Phone:    req.Phone,
		Address:  req.Address,
	}

	if err := h.userService.SignUp(c.Context(), userEntity); err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(response.DefaultResponse{
		Message: "user registered successfully",
		Data:    nil,
	})
}

func (h *userHandler) SignIn(c fiber.Ctx) error {
	var req request.SignInRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Warn().Err(err).Msg("failed to bind signin request")
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, token, err := h.userService.SignIn(c.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	userResp := &response.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Address:   user.Address,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "signin successful",
		Data: response.SignInResponse{
			Token: token,
			User:  userResp,
		},
	})
}

func (h *userHandler) GetProfile(c fiber.Ctx) error {
	userData, ok := c.Locals("user").(entity.JwtUserData)
	if !ok || userData.ID == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized access")
	}

	user, err := h.userService.GetProfile(c.Context(), userData.ID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user profile not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	userResp := response.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Address:   user.Address,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "get profile successful",
		Data:    userResp,
	})
}

func (h *userHandler) UpdateProfile(c fiber.Ctx) error {
	userData, ok := c.Locals("user").(entity.JwtUserData)
	if !ok || userData.ID == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized access")
	}

	var req request.UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	userEntity := entity.UserEntity{
		ID:       userData.ID,
		Name:     req.Name,
		Password: req.Password,
		Phone:    req.Phone,
		Address:  req.Address,
	}

	if err := h.userService.UpdateProfile(c.Context(), userEntity); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "profile updated successfully",
		Data:    nil,
	})
}

func (h *userHandler) DeleteUser(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := conv.StringToInt64(idParam)
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id parameter")
	}

	if err := h.userService.DeleteUser(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(response.DefaultResponse{
		Message: "user deleted successfully",
		Data:    nil,
	})
}

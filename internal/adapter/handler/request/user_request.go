package request

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type SignUpRequest struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Phone    string `json:"phone" validate:"omitempty,min=10,max=15"`
	Address  string `json:"address" validate:"omitempty"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name" validate:"omitempty,min=3"`
	Password string `json:"password" validate:"omitempty,min=6"`
	Phone    string `json:"phone" validate:"omitempty,min=10,max=15"`
	Address  string `json:"address" validate:"omitempty"`
}

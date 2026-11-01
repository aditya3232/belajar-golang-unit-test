package response

type UserResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type SignInResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

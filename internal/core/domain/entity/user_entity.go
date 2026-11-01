package entity

type UserEntity struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	Role      string `json:"role"`
	Token     string `json:"token,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type QueryStringUser struct {
	Search    string `json:"search"`
	Page      int64  `json:"page"`
	Limit     int64  `json:"limit"`
	OrderBy   string `json:"order_by"`
	OrderType string `json:"order_type"`
}

package dto

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
	Role     string `json:"role" binding:"required,oneof=ADMIN CASHIER"`
	Gender   string `json:"gender"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Role     string `json:"role" binding:"omitempty,oneof=ADMIN CASHIER"`
	Password string `json:"password"`
	Gender   string `json:"gender"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Gender    string `json:"gender"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

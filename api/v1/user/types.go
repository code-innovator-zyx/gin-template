package user

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2025/10/30 下午5:29
* @Package:
 */
// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe" description:"用户名"`
	Password string `json:"password" binding:"required" example:"password123" description:"密码"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com" description:"邮箱"`
	Nickname string `json:"nickname" example:"John Doe" description:"昵称"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe" description:"用户名"`
	Password string `json:"password" binding:"required" example:"password123" description:"密码"`
}

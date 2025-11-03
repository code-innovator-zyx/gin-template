package rbac

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/10/31 下午4:40
* @Package:
 */

// UserRegisterRequest 用户注册请求
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe" description:"用户名"`
	Password string `json:"password" binding:"required" example:"password123" description:"密码"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com" description:"邮箱"`
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Username string `json:"username" binding:"required" example:"johndoe" description:"用户名"`
	Password string `json:"password" binding:"required" example:"password123" description:"密码"`
}

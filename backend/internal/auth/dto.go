package auth

type AdminLoginReq struct {
	Password string `json:"password" form:"password" binding:"required"`
}

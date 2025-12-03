package rbac

import "time"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/2 下午4:56
* @Package:
 */

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id" example:"1" description:"ID"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" description:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z" description:"更新时间"`
}

func (BaseModel) TableName() string {
	panic("implement me")
}
func (b BaseModel) GetID() uint {
	return b.ID
}

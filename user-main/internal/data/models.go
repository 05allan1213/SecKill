package data

import "time"

type User struct {
	UserID     string `gorm:"column:id"`
	UserName   string
	Pwd        string
	Sex        int
	Age        int
	Email      string
	Contact    string
	Mobile     string
	IdCard     string
	CreateTime time.Time  `gorm:"column:create_time;default:null"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:null"`
}

func (p *User) TableName() string {
	return "t_user_info"
}

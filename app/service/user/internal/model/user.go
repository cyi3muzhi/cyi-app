package model

import (
	cgorm "cyi-app/pkg/util/gorm"
)

type User struct {
	cgorm.Model
	Name     string `gorm:"column:name"`
	Phone    string `gorm:"column:phone"`
	Password string `gorm:"column:password"`
	Username string `gorm:"column:username"`
}

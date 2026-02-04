package models

import (
	"cyi-app/pkg/util/gorm"
)

type Order struct {
	*gorm.Model
}

type IOrder interface {
}

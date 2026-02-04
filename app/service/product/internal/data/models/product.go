package models

import (
	"cyi-app/pkg/util/gorm"
)

type Product struct {
	*gorm.Model
}

type IProduct interface {
}

package models

import (
	"cyi-app/pkg/util/gorm"
)

type Task struct {
	*gorm.Model
}

type ITask interface {
}

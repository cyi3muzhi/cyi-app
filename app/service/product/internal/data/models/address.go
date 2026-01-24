package models

import cgorm "cyi-app/pkg/util/gorm"

type Address struct {
	cgorm.Model
	Address string `gorm:"column:address"`
}

type AddressRepo[T any] interface{}

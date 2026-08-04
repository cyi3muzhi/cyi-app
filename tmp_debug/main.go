package main

import (
	"context"
	"fmt"

	"cyi-app/app/service/user/internal/data/curd"
	mod "cyi-app/app/service/user/internal/data/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:mysql0411@tcp(127.0.0.1:3306)/cin_test?parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("open err:", err)
		return
	}
	tx := db.Begin()
	defer tx.Rollback()

	q := curd.Use(tx)
	info, err := q.User.WithContext(context.Background()).Updates(&mod.User{
		ID:       14,
		Nickname: "cyi用户",
		Email:    "1249236090@qq.com",
		Gender:   0,
		Birthday: "2000-02-28",
	})
	fmt.Printf("err=%v RowsAffected=%d\n", err, info.RowsAffected)
}

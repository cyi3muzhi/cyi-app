package gorminterface

import mod "cyi-app/app/service/user/internal/model"

type Query[T any] interface {
	// GetByID 根据 ID 进行查询
	// SELECT * FROM @@table WHERE id=@id
	GetByID(id int) (T, error)

	// FilterWithColumn 根据列名和值进行查询
	// SELECT * FROM @@table WHERE @@column=@value
	FilterWithColumn(column string, value string) (T, error)

	// SELECT * FROM @@table
	//   {{if user.ID > 0}}
	//       WHERE id=@user.ID
	//   {{else if user.Name != ""}}
	//       WHERE username=@user.Name
	//   {{else if user.Phone != ""}}
	//       WHERE phone=@user.Phone
	//   {{else if user.Username != ""}}
	//       WHERE username=@user.Username
	//   {{end}}
	QueryWith(user mod.User) (T, error)

	// UPDATE @@table
	//  {{set}}
	//    {{if user.Name != ""}} username=@user.Name, {{end}}
	//    {{if user.Age > 0}} age=@user.Age, {{end}}
	//    {{if user.Age >= 18}} is_adult=1 {{else}} is_adult=0 {{end}}
	//  {{end}}
	// WHERE id=@id
	// Update(user biz.User, id int) error

	// where("name=@name")
	// FilterByNameAnd(name string)

	// SELECT * FROM @@table
	//  {{where}}
	//    {{if !start.IsZero()}}
	//      created_time > @start
	//    {{end}}
	//    {{if !end.IsZero()}}
	//      AND created_time < @end
	//    {{end}}
	//  {{end}}
	// FilterWithTime(start, end time.Time) ([]T, error)
}

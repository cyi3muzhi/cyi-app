package curd

import (
	"context"
	"cyi-app/app/service/marketing/internal/data/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

func newCoupon(db *gorm.DB) coupon {
	_coupon := coupon{}
	_coupon.couponDo.UseDB(db)
	_coupon.couponDo.UseModel(&models.Coupon{})

	tableName := _coupon.couponDo.TableName()
	_coupon.ALL = field.NewField(tableName, "*")
	_coupon.ID = field.NewInt64(tableName, "id")
	_coupon.CreateTime = field.NewInt64(tableName, "create_time")
	_coupon.UpdateTime = field.NewInt64(tableName, "update_time")
	_coupon.DeletedAt = field.NewField(tableName, "deleted_at")
	_coupon.Name = field.NewString(tableName, "name")
	_coupon.Type = field.NewInt32(tableName, "type")
	_coupon.Value = field.NewInt64(tableName, "value")
	_coupon.MinAmount = field.NewInt64(tableName, "min_amount")
	_coupon.Quantity = field.NewInt32(tableName, "quantity")
	_coupon.IssuedCount = field.NewInt32(tableName, "issued_count")
	_coupon.UsedCount = field.NewInt32(tableName, "used_count")
	_coupon.ValidStart = field.NewInt64(tableName, "valid_start")
	_coupon.ValidEnd = field.NewInt64(tableName, "valid_end")
	_coupon.Status = field.NewInt32(tableName, "status")
	_coupon.Enabled = field.NewBool(tableName, "enabled")

	_coupon.fillFieldMap()
	return _coupon
}

type coupon struct {
	couponDo

	ALL          field.Field
	ID           field.Int64
	CreateTime   field.Int64
	UpdateTime   field.Int64
	DeletedAt    field.Field
	Name         field.String
	Type         field.Int32
	Value        field.Int64
	MinAmount    field.Int64
	Quantity     field.Int32
	IssuedCount  field.Int32
	UsedCount    field.Int32
	ValidStart   field.Int64
	ValidEnd     field.Int64
	Status       field.Int32
	Enabled      field.Bool

	fieldMap map[string]field.Expr
}

func (o coupon) Table(newTableName string) *coupon {
	o.couponDo.UseTable(newTableName)
	return o.updateTableName(newTableName)
}

func (o coupon) As(alias string) *coupon {
	o.couponDo.DO = *(o.couponDo.As(alias).(*gen.DO))
	return o.updateTableName(alias)
}

func (o *coupon) updateTableName(table string) *coupon {
	o.ALL = field.NewField(table, "*")
	o.ID = field.NewInt64(table, "id")
	o.CreateTime = field.NewInt64(table, "create_time")
	o.UpdateTime = field.NewInt64(table, "update_time")
	o.DeletedAt = field.NewField(table, "deleted_at")
	o.Name = field.NewString(table, "name")
	o.Type = field.NewInt32(table, "type")
	o.Value = field.NewInt64(table, "value")
	o.MinAmount = field.NewInt64(table, "min_amount")
	o.Quantity = field.NewInt32(table, "quantity")
	o.IssuedCount = field.NewInt32(table, "issued_count")
	o.UsedCount = field.NewInt32(table, "used_count")
	o.ValidStart = field.NewInt64(table, "valid_start")
	o.ValidEnd = field.NewInt64(table, "valid_end")
	o.Status = field.NewInt32(table, "status")
	o.Enabled = field.NewBool(table, "enabled")

	o.fillFieldMap()
	return o
}

func (o *coupon) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := o.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (o *coupon) fillFieldMap() {
	o.fieldMap = make(map[string]field.Expr, 14)
	o.fieldMap["id"] = o.ID
	o.fieldMap["create_time"] = o.CreateTime
	o.fieldMap["update_time"] = o.UpdateTime
	o.fieldMap["deleted_at"] = o.DeletedAt
	o.fieldMap["name"] = o.Name
	o.fieldMap["type"] = o.Type
	o.fieldMap["value"] = o.Value
	o.fieldMap["min_amount"] = o.MinAmount
	o.fieldMap["quantity"] = o.Quantity
	o.fieldMap["issued_count"] = o.IssuedCount
	o.fieldMap["used_count"] = o.UsedCount
	o.fieldMap["valid_start"] = o.ValidStart
	o.fieldMap["valid_end"] = o.ValidEnd
	o.fieldMap["status"] = o.Status
	o.fieldMap["enabled"] = o.Enabled
}

func (o coupon) clone(db *gorm.DB) coupon {
	o.couponDo.ReplaceDB(db)
	return o
}

type couponDo struct{ gen.DO }

func (o couponDo) Debug() *couponDo {
	return o.withDO(o.DO.Debug())
}

func (o couponDo) WithContext(ctx context.Context) *couponDo {
	return o.withDO(o.DO.WithContext(ctx))
}

func (o couponDo) Clauses(conds ...clause.Expression) *couponDo {
	return o.withDO(o.DO.Clauses(conds...))
}

func (o couponDo) Returning(value interface{}, columns ...string) *couponDo {
	return o.withDO(o.DO.Returning(value, columns...))
}

func (o couponDo) Not(conds ...gen.Condition) *couponDo {
	return o.withDO(o.DO.Not(conds...))
}

func (o couponDo) Or(conds ...gen.Condition) *couponDo {
	return o.withDO(o.DO.Or(conds...))
}

func (o couponDo) Select(conds ...field.Expr) *couponDo {
	return o.withDO(o.DO.Select(conds...))
}

func (o couponDo) Where(conds ...gen.Condition) *couponDo {
	return o.withDO(o.DO.Where(conds...))
}

func (o couponDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) *couponDo {
	return o.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (o couponDo) Order(conds ...field.Expr) *couponDo {
	return o.withDO(o.DO.Order(conds...))
}

func (o couponDo) Distinct(cols ...field.Expr) *couponDo {
	return o.withDO(o.DO.Distinct(cols...))
}

func (o couponDo) Omit(cols ...field.Expr) *couponDo {
	return o.withDO(o.DO.Omit(cols...))
}

func (o couponDo) Join(table schema.Tabler, on ...field.Expr) *couponDo {
	return o.withDO(o.DO.Join(table, on...))
}

func (o couponDo) LeftJoin(table schema.Tabler, on ...field.Expr) *couponDo {
	return o.withDO(o.DO.LeftJoin(table, on...))
}

func (o couponDo) RightJoin(table schema.Tabler, on ...field.Expr) *couponDo {
	return o.withDO(o.DO.RightJoin(table, on...))
}

func (o couponDo) Group(cols ...field.Expr) *couponDo {
	return o.withDO(o.DO.Group(cols...))
}

func (o couponDo) Having(conds ...gen.Condition) *couponDo {
	return o.withDO(o.DO.Having(conds...))
}

func (o couponDo) Limit(limit int) *couponDo {
	return o.withDO(o.DO.Limit(limit))
}

func (o couponDo) Offset(offset int) *couponDo {
	return o.withDO(o.DO.Offset(offset))
}

func (o couponDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *couponDo {
	return o.withDO(o.DO.Scopes(funcs...))
}

func (o couponDo) Unscoped() *couponDo {
	return o.withDO(o.DO.Unscoped())
}

func (o couponDo) Pluck(field field.Expr, dest interface{}) error {
	return o.DO.Pluck(field, dest)
}

func (o couponDo) Count() (count int64, err error) {
	return o.DO.Count()
}

func (o couponDo) First() (*models.Coupon, error) {
	if result, err := o.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.Coupon), nil
	}
}

func (o couponDo) Take() (*models.Coupon, error) {
	if result, err := o.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.Coupon), nil
	}
}

func (o couponDo) Last() (*models.Coupon, error) {
	if result, err := o.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.Coupon), nil
	}
}

func (o couponDo) Find() ([]*models.Coupon, error) {
	result, err := o.DO.Find()
	return result.([]*models.Coupon), err
}

func (o couponDo) FindInBatches(result *[]*models.Coupon, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return o.DO.FindInBatches(result, batchSize, fc)
}

func (o couponDo) FindByPage(offset int, limit int) (result []*models.Coupon, count int64, err error) {
	result, err = o.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}
	count, err = o.Count()
	return
}

func (o couponDo) Create(values ...*models.Coupon) error {
	if len(values) == 0 {
		return nil
	}
	return o.DO.Create(values)
}

func (o couponDo) CreateInBatches(values []*models.Coupon, batchSize int) error {
	return o.DO.CreateInBatches(values, batchSize)
}

func (o couponDo) Save(values ...*models.Coupon) error {
	if len(values) == 0 {
		return nil
	}
	return o.DO.Save(values)
}

func (o couponDo) FirstOrInit() (*models.Coupon, error) {
	if result, err := o.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.Coupon), nil
	}
}

func (o couponDo) FirstOrCreate() (*models.Coupon, error) {
	if result, err := o.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.Coupon), nil
	}
}

func (o couponDo) Update(column field.Expr, value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.Update(column, value)
}

func (o couponDo) Updates(value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.Updates(value)
}

func (o couponDo) UpdateColumn(column field.Expr, value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.UpdateColumn(column, value)
}

func (o couponDo) UpdateColumns(value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.UpdateColumns(value)
}

func (o couponDo) Delete() (result gen.ResultInfo, err error) {
	return o.DO.Delete()
}

func (o *couponDo) withDO(do gen.Dao) *couponDo {
	o.DO = *do.(*gen.DO)
	return o
}
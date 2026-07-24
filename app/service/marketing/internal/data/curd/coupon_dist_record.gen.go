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

func newCouponDistRecord(db *gorm.DB) couponDistRecord {
	_couponDistRecord := couponDistRecord{}
	_couponDistRecord.couponDistRecordDo.UseDB(db)
	_couponDistRecord.couponDistRecordDo.UseModel(&models.CouponDistRecord{})

	tableName := _couponDistRecord.couponDistRecordDo.TableName()
	_couponDistRecord.ALL = field.NewField(tableName, "*")
	_couponDistRecord.ID = field.NewInt64(tableName, "id")
	_couponDistRecord.CreateTime = field.NewInt64(tableName, "create_time")
	_couponDistRecord.UpdateTime = field.NewInt64(tableName, "update_time")
	_couponDistRecord.DeletedAt = field.NewField(tableName, "deleted_at")
	_couponDistRecord.CouponID = field.NewInt64(tableName, "coupon_id")
	_couponDistRecord.UserID = field.NewInt64(tableName, "user_id")
	_couponDistRecord.Source = field.NewString(tableName, "source")
	_couponDistRecord.Status = field.NewInt32(tableName, "status")
	_couponDistRecord.UsedTime = field.NewInt64(tableName, "used_time")
	_couponDistRecord.OrderID = field.NewInt64(tableName, "order_id")

	_couponDistRecord.fillFieldMap()
	return _couponDistRecord
}

type couponDistRecord struct {
	couponDistRecordDo

	ALL          field.Field
	ID           field.Int64
	CreateTime   field.Int64
	UpdateTime   field.Int64
	DeletedAt    field.Field
	CouponID     field.Int64
	UserID       field.Int64
	Source       field.String
	Status       field.Int32
	UsedTime     field.Int64
	OrderID      field.Int64

	fieldMap map[string]field.Expr
}

func (o couponDistRecord) Table(newTableName string) *couponDistRecord {
	o.couponDistRecordDo.UseTable(newTableName)
	return o.updateTableName(newTableName)
}

func (o couponDistRecord) As(alias string) *couponDistRecord {
	o.couponDistRecordDo.DO = *(o.couponDistRecordDo.As(alias).(*gen.DO))
	return o.updateTableName(alias)
}

func (o *couponDistRecord) updateTableName(table string) *couponDistRecord {
	o.ALL = field.NewField(table, "*")
	o.ID = field.NewInt64(table, "id")
	o.CreateTime = field.NewInt64(table, "create_time")
	o.UpdateTime = field.NewInt64(table, "update_time")
	o.DeletedAt = field.NewField(table, "deleted_at")
	o.CouponID = field.NewInt64(table, "coupon_id")
	o.UserID = field.NewInt64(table, "user_id")
	o.Source = field.NewString(table, "source")
	o.Status = field.NewInt32(table, "status")
	o.UsedTime = field.NewInt64(table, "used_time")
	o.OrderID = field.NewInt64(table, "order_id")

	o.fillFieldMap()
	return o
}

func (o *couponDistRecord) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := o.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (o *couponDistRecord) fillFieldMap() {
	o.fieldMap = make(map[string]field.Expr, 11)
	o.fieldMap["id"] = o.ID
	o.fieldMap["create_time"] = o.CreateTime
	o.fieldMap["update_time"] = o.UpdateTime
	o.fieldMap["deleted_at"] = o.DeletedAt
	o.fieldMap["coupon_id"] = o.CouponID
	o.fieldMap["user_id"] = o.UserID
	o.fieldMap["source"] = o.Source
	o.fieldMap["status"] = o.Status
	o.fieldMap["used_time"] = o.UsedTime
	o.fieldMap["order_id"] = o.OrderID
}

func (o couponDistRecord) clone(db *gorm.DB) couponDistRecord {
	o.couponDistRecordDo.ReplaceDB(db)
	return o
}

type couponDistRecordDo struct{ gen.DO }

func (o couponDistRecordDo) Debug() *couponDistRecordDo {
	return o.withDO(o.DO.Debug())
}

func (o couponDistRecordDo) WithContext(ctx context.Context) *couponDistRecordDo {
	return o.withDO(o.DO.WithContext(ctx))
}

func (o couponDistRecordDo) Clauses(conds ...clause.Expression) *couponDistRecordDo {
	return o.withDO(o.DO.Clauses(conds...))
}

func (o couponDistRecordDo) Returning(value interface{}, columns ...string) *couponDistRecordDo {
	return o.withDO(o.DO.Returning(value, columns...))
}

func (o couponDistRecordDo) Not(conds ...gen.Condition) *couponDistRecordDo {
	return o.withDO(o.DO.Not(conds...))
}

func (o couponDistRecordDo) Or(conds ...gen.Condition) *couponDistRecordDo {
	return o.withDO(o.DO.Or(conds...))
}

func (o couponDistRecordDo) Select(conds ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.Select(conds...))
}

func (o couponDistRecordDo) Where(conds ...gen.Condition) *couponDistRecordDo {
	return o.withDO(o.DO.Where(conds...))
}

func (o couponDistRecordDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) *couponDistRecordDo {
	return o.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (o couponDistRecordDo) Order(conds ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.Order(conds...))
}

func (o couponDistRecordDo) Distinct(cols ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.Distinct(cols...))
}

func (o couponDistRecordDo) Omit(cols ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.Omit(cols...))
}

func (o couponDistRecordDo) Join(table schema.Tabler, on ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.Join(table, on...))
}

func (o couponDistRecordDo) LeftJoin(table schema.Tabler, on ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.LeftJoin(table, on...))
}

func (o couponDistRecordDo) RightJoin(table schema.Tabler, on ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.RightJoin(table, on...))
}

func (o couponDistRecordDo) Group(cols ...field.Expr) *couponDistRecordDo {
	return o.withDO(o.DO.Group(cols...))
}

func (o couponDistRecordDo) Having(conds ...gen.Condition) *couponDistRecordDo {
	return o.withDO(o.DO.Having(conds...))
}

func (o couponDistRecordDo) Limit(limit int) *couponDistRecordDo {
	return o.withDO(o.DO.Limit(limit))
}

func (o couponDistRecordDo) Offset(offset int) *couponDistRecordDo {
	return o.withDO(o.DO.Offset(offset))
}

func (o couponDistRecordDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *couponDistRecordDo {
	return o.withDO(o.DO.Scopes(funcs...))
}

func (o couponDistRecordDo) Unscoped() *couponDistRecordDo {
	return o.withDO(o.DO.Unscoped())
}

func (o couponDistRecordDo) Pluck(field field.Expr, dest interface{}) error {
	return o.DO.Pluck(field, dest)
}

func (o couponDistRecordDo) Count() (count int64, err error) {
	return o.DO.Count()
}

func (o couponDistRecordDo) First() (*models.CouponDistRecord, error) {
	if result, err := o.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.CouponDistRecord), nil
	}
}

func (o couponDistRecordDo) Take() (*models.CouponDistRecord, error) {
	if result, err := o.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.CouponDistRecord), nil
	}
}

func (o couponDistRecordDo) Last() (*models.CouponDistRecord, error) {
	if result, err := o.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.CouponDistRecord), nil
	}
}

func (o couponDistRecordDo) Find() ([]*models.CouponDistRecord, error) {
	result, err := o.DO.Find()
	return result.([]*models.CouponDistRecord), err
}

func (o couponDistRecordDo) FindInBatches(result *[]*models.CouponDistRecord, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return o.DO.FindInBatches(result, batchSize, fc)
}

func (o couponDistRecordDo) FindByPage(offset int, limit int) (result []*models.CouponDistRecord, count int64, err error) {
	result, err = o.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}
	count, err = o.Count()
	return
}

func (o couponDistRecordDo) Create(values ...*models.CouponDistRecord) error {
	if len(values) == 0 {
		return nil
	}
	return o.DO.Create(values)
}

func (o couponDistRecordDo) CreateInBatches(values []*models.CouponDistRecord, batchSize int) error {
	return o.DO.CreateInBatches(values, batchSize)
}

func (o couponDistRecordDo) Save(values ...*models.CouponDistRecord) error {
	if len(values) == 0 {
		return nil
	}
	return o.DO.Save(values)
}

func (o couponDistRecordDo) FirstOrInit() (*models.CouponDistRecord, error) {
	if result, err := o.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.CouponDistRecord), nil
	}
}

func (o couponDistRecordDo) FirstOrCreate() (*models.CouponDistRecord, error) {
	if result, err := o.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.CouponDistRecord), nil
	}
}

func (o couponDistRecordDo) Update(column field.Expr, value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.Update(column, value)
}

func (o couponDistRecordDo) Updates(value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.Updates(value)
}

func (o couponDistRecordDo) UpdateColumn(column field.Expr, value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.UpdateColumn(column, value)
}

func (o couponDistRecordDo) UpdateColumns(value interface{}) (result gen.ResultInfo, err error) {
	return o.DO.UpdateColumns(value)
}

func (o couponDistRecordDo) Delete() (result gen.ResultInfo, err error) {
	return o.DO.Delete()
}

func (o *couponDistRecordDo) withDO(do gen.Dao) *couponDistRecordDo {
	o.DO = *do.(*gen.DO)
	return o
}
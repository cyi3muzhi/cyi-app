package biz

// Category 商品分类业务实体（纯领域模型，无 gorm/proto tag）
type Category struct {
	ID          int64
	ParentID    int64
	Name        string
	Slug        string
	Icon        string
	Image       string
	Sort        int32
	Status      bool
	Showcase    bool
	Count       int32
	Level       int32
	SeoTitle    string
	SeoKeywords string
	SeoDesc     string
	Children    []*Category
}

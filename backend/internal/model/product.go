package model

import (
	"time"
)

type AllergenType string
type UnitType string

const (
	AllergenNone       AllergenType = "none"       // 无过敏原
	AllergenGluten     AllergenType = "gluten"     // 含麸质谷物
	AllergenCrustacean AllergenType = "crustacean" // 甲壳类
	AllergenEgg        AllergenType = "egg"        // 蛋类
	AllergenFish       AllergenType = "fish"       // 鱼类
	AllergenPeanut     AllergenType = "peanut"     // 花生
	AllergenSoy        AllergenType = "soy"        // 大豆
	AllergenMilk       AllergenType = "milk"       // 乳制品
	AllergenNut        AllergenType = "nut"        // 坚果
	AllergenCelery     AllergenType = "celery"     // 芹菜
	AllergenMustard    AllergenType = "mustard"    // 芥末
	AllergenSesame     AllergenType = "sesame"     // 芝麻
	AllergenSulphite   AllergenType = "sulphite"   // 亚硫酸盐
	AllergenLupin      AllergenType = "lupin"      // 羽扇豆
	AllergenMollusc    AllergenType = "mollusc"    // 软体动物

	UnitGram       UnitType = "gram"
	UnitKilogram   UnitType = "kilogram"
	UnitMilliliter UnitType = "milliliter"
	UnitLiter      UnitType = "liter"
	UnitPiece      UnitType = "piece"
	UnitBowl       UnitType = "bowl"
	UnitServing    UnitType = "serving"
)

type Product struct {
	ID uint64 `gorm:"primaryKey"`
	// 商品代码,例如"康师傅红烧牛肉面","海底捞牛肉面"等都属于"面"或者"牛肉面"这个食材的商品
	ProductCode string `gorm:"not null;unique"`
	// 食材代码代表该商品是哪种食材,例如"面条","猪肉"等,是一个大的品类,因为不同品牌的商品可以是同一种食材
	IngredientCode string `gorm:"not null"`
	// 商品名称,例如"康师傅红烧牛肉面","海底捞牛肉面"等
	Name string `gorm:"not null"`
	// 商品单位量,例如"100.00"
	Amount float64 `gorm:"type:decimal(6,2);not null"`
	// 商品单位,例如"g","kg","ml","l","item"等
	Unit UnitType `gorm:"not null"`
	// 商品描述,例如"这是一个100g的康师傅红烧牛肉面"
	Description string `gorm:"type:text"`
	// 商品价格,例如"100.00"
	Price float64 `gorm:"decimal(10,2);not null"` // 价格
	// 过敏原一个由商品控制而非食材,因为即使同一种食材(例如面条),也可能因为不同品牌而有不同的过敏原
	AllergenType AllergenType `gorm:"type:varchar(20)"` // 过敏原类型
	CreatedAt    time.Time    `gorm:"not null;default:current_timestamp"`
	UpdatedAt    time.Time    `gorm:"not null;default:current_timestamp"`
	//DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (p *Product) GetID() uint64 {
	return p.ID
}

func (p *Product) PrimaryKey() string {
	return "id"
}

func (p *Product) TableName() string {
	return "products"
}

package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Appearance struct {
	gorm.Model
	Address         string `gorm:"index; not null"`
	BlockNumber     uint32 `gorm:"not null"`
	TransactionId   uint32 `gorm:"not null"`
	BlockRangeStart uint64 `gorm:"not null"`
	BlockRangeEnd   uint64 `gorm:"not null"`
	// AppearanceId is unique and has format of "address.block_number.transaction_id"
	AppearanceId string `gorm:"index;unique;not null"`
}

func (a *Appearance) BeforeSave(tx *gorm.DB) (err error) {
	a.SetAppearanceId()
	return
}

func (a *Appearance) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.AddClause(clause.OnConflict{DoNothing: true})
	return nil
}

func (a *Appearance) SetAppearanceId() {
	a.AppearanceId = fmt.Sprintf("%s.%d.%d", a.Address, a.BlockNumber, a.TransactionId)
}

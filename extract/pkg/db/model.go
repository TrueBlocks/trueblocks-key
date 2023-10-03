package db

import "gorm.io/gorm"

type Appearance struct {
	gorm.Model
	Address         string `gorm:"index"`
	BlockNumber     uint32
	TransactionId   uint32
	BlockRangeStart uint64
	BlockRangeEnd   uint64
}

type Progress struct {
	gorm.Model
	Version         string
	BlockRangeStart uint64
	BlockRangeEnd   uint64
}

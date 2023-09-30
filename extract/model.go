package main

import "gorm.io/gorm"

type Appearance struct {
	gorm.Model
	Address         string
	BlockNumber     uint32
	TransactionId   uint32
	BlockRangeStart uint64
	BlockRangeEnd   uint64
}

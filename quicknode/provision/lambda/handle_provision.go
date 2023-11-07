package main

import (
	"github.com/gin-gonic/gin"
)

func HandleProvision(c *gin.Context) {
	upsertAccount(c, false)
}

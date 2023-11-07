package main

import (
	"github.com/gin-gonic/gin"
)

func HandleUpdate(c *gin.Context) {
	upsertAccount(c, true)
}

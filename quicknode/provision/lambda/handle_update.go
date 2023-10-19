package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleUpdate(c *gin.Context) {
	var account *Account

	err := c.BindJSON(&account)
	if err != nil {
		log.Println("update: binding account:", err)
		c.AbortWithError(http.StatusBadRequest, errors.New("could not parse JSON"))
		return
	}

	// Read account data
	dbItem, err := account.DynamoGet()
	if err != nil {
		log.Println("update: account.DynamoRead:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}
	if dbItem == nil {
		// This account is not registered
		log.Println("update: account not found", account.QuicknodeId)
		c.AbortWithError(http.StatusNotFound, errors.New("account not found"))
		return
	}

	// Save updated account
	if err = account.DynamoPut(); err != nil {
		log.Println("update: saving account", account.QuicknodeId, ":", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

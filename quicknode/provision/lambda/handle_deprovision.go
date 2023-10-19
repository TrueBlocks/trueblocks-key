package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleDeprovision(c *gin.Context) {
	var account *Account

	err := c.BindJSON(&account)
	if err != nil {
		log.Println("deprovision: binding account:", err)
		c.AbortWithError(http.StatusBadRequest, errors.New("could not parse JSON"))
		return
	}

	// Read account data
	dbItem, err := account.DynamoGet()
	if err != nil {
		log.Println("deprovision: account.DynamoRead:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}
	if dbItem == nil {
		// This account is not registered
		log.Println("deprovision: account not found", account.QuicknodeId)
		c.AbortWithError(http.StatusNotFound, errors.New("account not found"))
		return
	}

	// Remove the account
	if err = account.DynamoDelete(); err != nil {
		log.Println("deprovision: saving account", account.QuicknodeId, ":", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

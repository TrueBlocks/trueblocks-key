package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	qnaccount "trueblocks.io/quicknode/account"
)

func HandleDeactivateEndpoint(c *gin.Context) {
	// Since we don't need to support endpoint, we will just
	// validate the account and return success if it's registered
	account := qnaccount.NewAccount(svc, cnf.QnProvision.TableName)

	err := c.BindJSON(&account)
	if err != nil {
		log.Println("deactivate: binding account:", err)
		c.AbortWithError(http.StatusBadRequest, errors.New("could not parse JSON"))
		return
	}

	// Read account data
	result, err := account.DynamoGet()
	if err != nil {
		log.Println("deactivate: account.DynamoRead:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}
	if result == nil {
		// This account is not registered
		log.Println("deactivate: account not found", account.QuicknodeId)
		c.AbortWithError(http.StatusNotFound, errors.New("account not found"))
		return
	}

	// Do nothing
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

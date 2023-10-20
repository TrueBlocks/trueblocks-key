package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	qnaccount "trueblocks.io/quicknode/account"
)

func HandleProvision(c *gin.Context) {
	account := qnaccount.NewAccount(svc, cnf.QnProvision.TableName)
	success := func() {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}

	err := c.BindJSON(&account)
	if err != nil {
		log.Println("provision: binding account:", err)
		c.AbortWithError(http.StatusBadRequest, errors.New("could not parse JSON"))
		return
	}

	// First check if the user is already in the database
	result, err := account.DynamoGet()
	if err != nil {
		log.Println("provision: account.DynamoRead:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}
	if result != nil {
		// We already have the account registered
		log.Println("account already registered", account.QuicknodeId)
		success()
		return
	}

	log.Println("Adding new account", account.QuicknodeId)

	err = account.DynamoPut()
	if err != nil {
		log.Println("provision: account.DynamoPut:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}

	success()
	return
}

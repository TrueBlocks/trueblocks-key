package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	qnaccount "trueblocks.io/quicknode/account"
)

func HandleProvision(c *gin.Context) {
	account := qnaccount.NewAccount(dynamoClient, cnf.QnProvision.TableName)
	success := func() {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}

	err := c.BindJSON(account)
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

	if err = initApiGateway(); err != nil {
		log.Println("initApiGateway:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}
	apiKey, err := qnaccount.FindByPlanSlug(apiGatewayClient, account.Plan)
	if err != nil {
		log.Println("fetching API key for plan", account.Plan, ":", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}
	account.ApiKey = *apiKey

	err = account.DynamoPut()
	if err != nil {
		log.Println("provision: account.DynamoPut:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}

	success()
	return
}

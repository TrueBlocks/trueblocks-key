package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	qnaccount "trueblocks.io/quicknode/account"
)

func upsertAccount(c *gin.Context, mustExist bool) {
	success := func() {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	}

	accountData, err := qnaccount.NewAccountData(c)
	if err != nil {
		log.Println("parsing account data:", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	account := qnaccount.NewAccount(dynamoClient, cnf.QnProvision.TableName)
	account.QuicknodeId = accountData.QuicknodeId

	// First check if the user is already in the database
	result, err := account.DynamoGet()
	if err != nil {
		log.Println("provision: account.DynamoRead:", err)
		c.AbortWithError(http.StatusInternalServerError, nil)
		return
	}
	if result != nil && account.HasEndpointId(accountData.EndpointId) {
		// We already have the account registered
		log.Println("account already registered", account.QuicknodeId)
		success()
		return
	}

	if !mustExist {
		if result == nil {
			log.Println("Adding new account", account.QuicknodeId)
		} else {
			log.Println("Registering new endpoint", accountData.EndpointId, "for account", account.QuicknodeId)
		}
	}

	if mustExist && result == nil {
		log.Println("account not found", account.QuicknodeId)
		c.AbortWithError(http.StatusNotFound, errors.New("account not found"))
		return
	}

	account.SetFromAccountData(accountData)

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

package main

import (
	"log"
	"net/http"

	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/gin-gonic/gin"
)

func HandleUpdate(c *gin.Context) {
	resp := &responder{c}
	log.Println("Update account")

	account, accountData, err := readAccountFromRequest(resp, c)
	if err != nil {
		return
	}
	account.EndpointIds = append(account.EndpointIds, accountData.EndpointId)

	oldAccount := qnaccount.NewAccount(dynamoClient, cnf.QnProvision.TableName)
	oldAccount.SetFromAccountData(accountData)
	oldAccount.EndpointIds = append(oldAccount.EndpointIds, accountData.EndpointId)
	if err := findAccount(resp, oldAccount); err != nil {
		return
	}

	log.Println("Account:", account.QuicknodeId, "old plan:", oldAccount.Plan, "new plan:", account.Plan)

	if err := account.Authorize(accountData); err != nil {
		log.Println(err.Error(), accountData.QuicknodeId, accountData.EndpointId)
		resp.abortWithError(http.StatusNotFound, err)
		return
	}

	// load possibly new API key
	if err := account.LoadApiKey(apiGatewayClient); err != nil {
		log.Println(err)
		resp.abortWithInternalError()
		return
	}

	err = account.DynamoPut()
	if err != nil {
		log.Println(err)
		resp.abortWithInternalError()
		return
	}
	resp.success()
}

func findAccount(resp *responder, account *qnaccount.Account) (err error) {
	found, err := account.Find()
	if err != nil {
		log.Println(err)
		resp.abortWithInternalError()
		return
	}
	if !found {
		log.Println("account not found", account.QuicknodeId)
		resp.abortWithCode(http.StatusNotFound)
		return
	}
	return
}

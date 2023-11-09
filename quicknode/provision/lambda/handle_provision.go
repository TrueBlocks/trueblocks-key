package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	qnaccount "trueblocks.io/quicknode/account"
)

func HandleProvision(c *gin.Context) {
	resp := &responder{c}

	account, accountData, err := readAccountFromRequest(resp, c)
	if err != nil {
		return
	}

	found, _ := account.Find()
	if !found {
		log.Println("creating new account:", account.QuicknodeId, accountData.EndpointId)
	} else {
		log.Println("account exists, adding new endpoint", account.QuicknodeId, accountData.EndpointId)
	}

	if accountData.EndpointId == "" {
		log.Println("empty EndpointId")
		resp.abortWithError(http.StatusBadRequest, errors.New("missing endpoint ID"))
		return
	}

	if eid := accountData.EndpointId; !account.HasEndpointId(eid) {
		log.Println("Adding new endpoint", eid)
		account.ActivateEndpoint(eid)
	}

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

func readAccountFromRequest(resp *responder, c *gin.Context) (account *qnaccount.Account, accountData *qnaccount.AccountData, err error) {
	accountData, err = qnaccount.NewAccountData(c)
	if err != nil {
		log.Println("parsing account data:", err)

		if err != qnaccount.ErrInvalidChainNetwork {
			resp.abortWithInvalidJson()
		} else {
			resp.abortWithError(http.StatusBadRequest, err)
		}
		return
	}
	account = qnaccount.NewAccount(dynamoClient, cnf.QnProvision.TableName)
	account.SetFromAccountData(accountData)

	return
}

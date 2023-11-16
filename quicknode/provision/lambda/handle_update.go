package main

import (
	"log"
	"net/http"

	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/gin-gonic/gin"
)

func HandleUpdate(c *gin.Context) {
	resp := &responder{c}

	account, accountData, err := readAccountFromRequest(resp, c)
	if err != nil {
		return
	}

	if err := findAccount(resp, account); err != nil {
		return
	}

	if err := account.Authorize(accountData); err != nil {
		log.Println(err.Error(), accountData.QuicknodeId, accountData.EndpointId)
		resp.abortWithError(http.StatusNotFound, err)
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

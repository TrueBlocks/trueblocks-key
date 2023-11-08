package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func HandleDeprovision(c *gin.Context) {
	resp := &responder{c}

	account, _, err := readAccountFromRequest(resp, c)
	if err != nil {
		return
	}

	if err := findAccount(resp, account); err != nil {
		return
	}

	// Remove the account
	if err = account.DynamoDelete(); err != nil {
		log.Println("deprovision: saving account", account.QuicknodeId, ":", err)
		resp.abortWithInternalError()
		return
	}

	resp.success()
}

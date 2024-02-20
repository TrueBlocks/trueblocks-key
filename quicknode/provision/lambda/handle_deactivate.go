package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleDeactivateEndpoint(c *gin.Context) {
	resp := &responder{c}
	log.Println("Deactivating endpoint")

	account, accountData, err := readAccountFromRequest(resp, c)
	if err != nil {
		return
	}

	if err := findAccount(resp, account); err != nil {
		return
	}

	endpointFound := account.DeactivateEndpoint(accountData.EndpointId)
	log.Println("endpoint found:", endpointFound)
	if !endpointFound {
		resp.abortWithError(http.StatusNotFound, errors.New("endpoint not found"))
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

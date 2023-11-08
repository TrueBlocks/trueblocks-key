package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errInvalidJson = errors.New("invalid JSON")

type responder struct {
	c *gin.Context
}

func (r *responder) abortWithInternalError() {
	r.c.AbortWithStatus(http.StatusInternalServerError)
}

func (r *responder) abortWithInvalidJson() {
	r.abortWithError(http.StatusBadRequest, errInvalidJson)
}

func (r *responder) abortWithCode(code int) {
	r.abortWithError(code, errors.New(http.StatusText(code)))
}

func (r *responder) abortWithError(code int, err error) {
	r.c.AbortWithStatusJSON(code, map[string]string{
		"error": err.Error(),
	})
}

func (r *responder) success() {
	r.c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

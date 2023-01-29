package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

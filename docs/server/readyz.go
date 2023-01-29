package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func readyz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

package main

import (
	"github.com/gin-gonic/gin"
)

func SetEndpoints(r *gin.Engine) {
	// Inform kube we are healthy but not necessarily ready
	r.GET("/healthz", healthz)
	// Inform kube or other watcher that app is ready
	// this differs from healthz in that this means it
	// can reach things like the API which it may not
	// need to in tests etc
	r.GET("/readyz", readyz)
}

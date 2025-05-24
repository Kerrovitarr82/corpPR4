package main

import (
	"corpPR4/internal/transport"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()
	transport.InitRoutes(r)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Aman913k/url-shortner/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRoutes(router *gin.Engine) {
	router.GET("/:url", routes.ResolveURL)
	router.POST("/api/v1", routes.ShortenURL)
}

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	router := gin.Default()

	//router.Use(gin.Logger())

	setupRoutes(router)

	port := os.Getenv("APP_PORT")
	
	if port == "" {
		port = ":3000" 
	}

	err = router.Run(port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

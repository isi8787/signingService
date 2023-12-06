package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	listenAddress := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddress = ":" + val
	}
	getDatabase()

	go generatePaillierKeys()

	fmt.Printf("** Service Started on Port %s **", listenAddress)

	router := gin.Default()

	NewRouter(router)

	router.Use(CORSMiddleware())

	router.Run(listenAddress)

}

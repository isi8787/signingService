package main

import (
	"github.com/gin-gonic/gin"
)

// NewRouter list of api end points
func NewRouter(router *gin.Engine) {

	// websocket route for performing joint ecdsa signing operations with mobile client
	router.GET("/WSHome/:userId/:blockchainId/:accountName", func(c *gin.Context) {
		wsHandler(c)
	})

	//postEDDSASignature provides api endpoint for sending partial eddsa signature
	// and completing the signing and aggregation using the custody service
	router.POST("/api/postEDDSASignature/:userId/:blockchainId/:accountName", HandlerWrap(POSTEDDSASignature))

	//postShare provides api endpoint for saving key share data from a customer for
	// a specific blockchain that uses ecdsa signing algorithm
	router.POST("/api/postECDSAShare", HandlerWrap(PostECDSAKeyShare))

	router.POST("/api/postEDDSAShare", HandlerWrap(PostEDDSAKeyShare))

	//getShare provides api endpoint for retriving key share data from a customer for
	// a specific blockchain that uses ecdsa signing algorithm. This should only be accessible
	// after prolong verification process
	router.GET("/api/getECDSAShare/:userId/:blockchainId/:accountName", HandlerWrap(GetECDSAKeyShare))

	router.GET("/api/getEDDSAShare/:userId/:blockchainId/:accountName", HandlerWrap(GetEDDSAKeyShare))

	//getUserAccounts provides api endpoint for retriving a client accounts information.
	// This should only be accessible after prolong verification process
	router.GET("/api/getUserAccounts/:userId", HandlerWrap(GetUserAccounts))

	//recoverUserAccounts provides api endpoint for creating and updating the recovery record
	// to manage release of key shares
	router.POST("/api/recoverUserAccounts/:userId", HandlerWrap(RecoverUserAccounts))

	//recoverUserAccounts provides api endpoint for creating and updating the recovery record
	// to manage release of key shares
	router.GET("/api/recoverUserAccounts/:userId", HandlerWrap(GetRecoverUserAccountsStatus))

	//initiate request to generate paillier key
	router.POST("/api/requestPaillierKey/:userId", HandlerWrap(RequestPaillierKey))

	//initiate request to generate paillier key
	router.GET("/api/requestPaillierKey/:userId", HandlerWrap(GetPaillierKey))

	//initiate request to generate paillier key
	router.DELETE("/api/requestPaillierKey/:userId", HandlerWrap(RemovePaillierKey))

	//initiate request to get random paillier keys
	router.GET("/api/requestPaillierKey", HandlerWrap(GetRandomPaillierKey))

	return
}

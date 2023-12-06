package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	flowErrors "bitbucket.org/carsonliving/flow.packages.errors"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func WriteErrorResponse(statusCode int, message string, w http.ResponseWriter) {
	log.Error(message)
	// @TODO: need to integration flow error library to send valid type and message.
	errorResponse := ApiError{
		Status: false,
		Err: ErrorDetails{
			Type:    flowErrors.ErrorTypeMap[statusCode],
			Message: message,
		},
	}
	respJson, errRespJson := json.Marshal(errorResponse)
	if errRespJson != nil {
		WriteErrorResponse(flowErrors.JsonMarshalErrorCode, errRespJson.Error(), w)
		return
	}

	WriteCustomResponse(statusCode, respJson, w) //Returning 200 since error code is wrapped into response
}

// to write custom response to the request
func WriteCustomResponse(code int, res []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	WriteRawResponse(code, res, w)
}

// to validate the response body and write raw response from 3rd party API
func ValidateAndWriteResponse(resp interface{}, err error, w http.ResponseWriter) {
	if err != nil {
		WriteErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	successResponse := ApiSuccess{
		Status: true,
		Result: resp,
	}

	successResponseBytes, err := json.Marshal(successResponse)
	if err != nil {
		WriteErrorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	WriteRawResponse(http.StatusOK, successResponseBytes, w)
}

func WriteResponse(code int, res interface{}, w http.ResponseWriter) {
	if res == nil {
		w.Header().Set("Content-Type", "application/json")
		WriteRawResponse(code, []byte{}, w)
		return
	}
	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Marshal JSON response failed, error=%q\n", err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json")
		WriteRawResponse(code, b, w)
	}
}

func WriteRawResponse(code int, res []byte, w http.ResponseWriter) {
	w.WriteHeader(code)
	w.Write(res)
}

// Validate a request
// TODO need to add all request validation functions including authorization header
// content type, origin request
func HandlerWrap(f func(c *gin.Context)) gin.HandlerFunc {

	return func(c *gin.Context) {
		f(c)
	}
}

func HandlerWrapMux(hf http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*reqToken := r.Header.Get("Authorization") //Authorization: Bearer somecrazylongtokenthatsfartoolongtoread
		fmt.Printf("%+v\n", reqToken)
		splitToken := strings.Split(reqToken, "Bearer")
		if len(splitToken) != 2 {
			http.Error(w, "Token doesn't seem right", http.StatusUnauthorized)
		}

		reqToken = strings.TrimSpace(splitToken[1]) //I don't want the word Bearer.

		keySet, err := jwk.Fetch(r.Context(), "https://fincodev.b2clogin.com/fincodev.onmicrosoft.com/B2C_1_signing/discovery/v2.0/keys")
		fmt.Println("KEYSET: ", keySet, " ERR: ", err)

		token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwa.RS256.String() {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid header not found")
			}

			keys, ok := keySet.LookupKeyID(kid)
			if !ok {
				return nil, fmt.Errorf("key %v not found", kid)
			}

			publickey := &rsa.PublicKey{}
			err = keys.Raw(publickey)
			if err != nil {
				return nil, fmt.Errorf("could not parse pubkey")
			}

			return publickey, nil
		})

		fmt.Println("Token data:", token)

		fmt.Println("Some claim: ", token.Claims)

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			fmt.Println("User id", claims["azp"])
		} else {
			fmt.Println(err)
		}

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		*/
		hf(w, r)
	})
}

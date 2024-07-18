package parser

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v5"
)

func ParseJWT(tokenString string) string {
	// Parse the JWT token
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		parsed, err := json.Marshal(claims)
		if err != nil {
			return ""
		}
		return string(parsed)
	} else {
		return ""
	}
}

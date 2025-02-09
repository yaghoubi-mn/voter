package jwt

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey []byte

func Init(secretKeyIn []byte) {
	secretKey = secretKeyIn
}
func CreateJWT(mapClaims map[string]any) (string, error) {
	if _, ok := mapClaims["exp"]; !ok {
		return "", errors.New("exp not found in mapClaims")
	}

	jwtMapClaims := jwt.MapClaims{}
	for key := range jwtMapClaims {
		jwtMapClaims[key] = mapClaims[key]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtMapClaims)

	tokenString, err := token.SignedString(secretKey)

	return tokenString, err
}

func VerifyJWT(tokenString string) (mapClaims map[string]any, err error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// extract claims
	jwtMapClaims := token.Claims.(jwt.MapClaims)
	mapClaims = make(map[string]any)

	for key := range jwtMapClaims {

		mapClaims[key] = jwtMapClaims[key]
	}

	return mapClaims, nil
}

func CreateRefreshAndAccessFromUser(refreshExpireTime, accessExpireTime time.Duration, userId uint64, username string) (refresh, access string, err error) {
	refresh, err = CreateJWT(map[string]any{
		"exp": time.Now().Add(refreshExpireTime).Unix(),
	})
	if err != nil {
		return "", "", err
	}
	access, err = CreateAccessFromUser(accessExpireTime, userId, username)

	return refresh, access, err
}

func CreateAccessFromUser(accessExpireTime time.Duration, userId uint64, username string) (access string, err error) {
	if userId == 0 {
		return "", errors.New("cannot create jwt: id is zero")
	}
	access, err = CreateJWT(map[string]any{
		"exp":      time.Now().Add(accessExpireTime).Unix(),
		"userId":   userId,
		"username": username,
	})

	return access, err
}

func GetUserFromAccess(access string) (userId uint64, username string, err error) {
	mapClaims, err := VerifyJWT(access)
	if err != nil {
		return 0, "", err
	}

	userId = mapClaims["userId"].(uint64)
	username = mapClaims["username"].(string)

	return userId, username, nil
}

func CreateRefreshAndAccessFromUserWithMap(refreshExpireTime, accessExpireTime time.Duration, userId uint64, username string) (tokens map[string]string, err error) {
	tokens = make(map[string]string)
	refresh, access, err := CreateRefreshAndAccessFromUser(refreshExpireTime, accessExpireTime, userId, username)
	tokens["refresh"] = refresh
	tokens["access"] = access
	tokens["accessExpireSeconds"] = strconv.Itoa(int(accessExpireTime.Seconds()))

	return
}

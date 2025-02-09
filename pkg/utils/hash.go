package utils

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const saltSize = 16
const bcryptCost = bcrypt.DefaultCost

func GenerateRandomSalt() (string, error) {

	salt := make([]byte, saltSize)

	_, err := rand.Read(salt[:])

	return hex.EncodeToString(salt), err
}

func HashPasswordWithSalt(password string, salt string) (string, error) {
	passwordBytes := []byte(password + salt)
	hashBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcryptCost)

	return string(hashBytes), err
}

func CompareHashAndPassword(hashedPassword, password, salt string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

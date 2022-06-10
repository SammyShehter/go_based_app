package middlewares

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
	"time"
)

type ClaimsWithScope struct {
	jwt.StandardClaims
	Scope string
}

const SecretKey string = "secret"

func IsAuthenticated(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	token, err := jwt.ParseWithClaims(cookie, &ClaimsWithScope{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})

	if err != nil || !token.Valid {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Not Allowed",
		})
	}

	payload := token.Claims.(*ClaimsWithScope)
	isAmbassador := strings.Contains(c.Path(), "/api/ambassador")

	if payload.Scope == "admin" && isAmbassador || payload.Scope == "ambassador" && !isAmbassador {
		fmt.Println("Violation detected!\n %a", payload)
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Not Allowed",
		})
	}

	return c.Next()
}

func GenerateJWT(userId uint, scope string) (string, error) {
	payload := ClaimsWithScope{}
	payload.ExpiresAt = time.Now().Add(time.Hour * 24).Unix()
	payload.Subject = strconv.Itoa(int(userId))
	payload.Scope = scope

	return jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(SecretKey))
}

func GetUserId(c *fiber.Ctx) (uint, error) {
	cookie := c.Cookies("jwt")
	token, err := jwt.ParseWithClaims(cookie, &ClaimsWithScope{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SecretKey), nil
		})
	if err != nil {
		return 0, err
	}

	payload := token.Claims.(*ClaimsWithScope)

	id, _ := strconv.Atoi(payload.Subject)

	return uint(id), nil
}

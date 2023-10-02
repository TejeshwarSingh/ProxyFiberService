package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

func main() {
	app := fiber.New()

	// Read the target server URL from the environment variable
	targetServer := os.Getenv("TARGET_SERVER_URL")
	if targetServer == "" {
		log.Fatal("Environment variable TARGET_SERVER_URL is not set")
		return
	}

	// Read the encryption service URL from the environment variable
	encryptionServiceURL := os.Getenv("ENCRYPTION_SERVICE_URL")
	if encryptionServiceURL == "" {
		log.Fatal("Environment variable ENCRYPTION_SERVICE_URL is not set")
		return
	}

	// Middleware to capture, encrypt, and replace x-user-name header
	app.Use(func(c *fiber.Ctx) error {
		userName := c.Get("x-user-name")
		if userName != "" {
			encryptedUserName, err := encryptUserName(userName, encryptionServiceURL)
			if err != nil {
				return c.Status(500).SendString("Failed to encrypt user name")
			}
			c.Request().Header.Set("x-user-name", encryptedUserName)
		}
		return c.Next()
	})

	// Set up the reverse proxy middleware
	proxyConfig := proxy.Config{
		Servers: []string{
			targetServer,
		},
	}

	app.Use(proxy.Balancer(proxyConfig))

	log.Printf("Starting server on :3000 and proxying to %s\n", targetServer)
	log.Fatal(app.Listen(":3100"))
}

func encryptUserName(userName string, encryptionServiceURL string) (string, error) {
	payload := map[string]string{"userName": userName}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(encryptionServiceURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("encryption service responded with status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var responseMap map[string]string
	err = json.Unmarshal(body, &responseMap)
	if err != nil {
		return "", err
	}

	encryptedUserName, exists := responseMap["encryptedUserName"]
	if !exists {
		return "", fmt.Errorf("encrypted user name not found in response")
	}

	return encryptedUserName, nil
}

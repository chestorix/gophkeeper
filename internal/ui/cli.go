// internal/ui/cli.go
package ui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	_ "strings"
	"time"

	"github.com/chestorix/gophkeeper/internal/client"
	"github.com/chestorix/gophkeeper/internal/models"
)

type CLI struct {
	client  *client.Client
	config  *client.Config
	scanner *bufio.Scanner
}

func NewCLI(client *client.Client, config *client.Config) *CLI {
	return &CLI{
		client:  client,
		config:  config,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (c *CLI) Run() error {
	if err := os.MkdirAll(c.config.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	for {
		fmt.Println("\n=== GophKeeper Client ===")
		fmt.Println("1. Register")
		fmt.Println("2. Login")
		fmt.Println("3. Save data")
		fmt.Println("4. List data")
		fmt.Println("5. Get data")
		fmt.Println("6. Update data")
		fmt.Println("7. Delete data")
		fmt.Println("8. Sync data")
		fmt.Println("9. Health check")
		fmt.Println("0. Exit")
		fmt.Print("Choose an option: ")

		c.scanner.Scan()
		choice := c.scanner.Text()

		ctx := context.Background()

		switch choice {
		case "1":
			c.register(ctx)
		case "2":
			c.login(ctx)
		case "3":
			c.saveData(ctx)
		case "4":
			c.listData(ctx)
		case "5":
			c.getData(ctx)
		case "6":
			c.updateData(ctx)
		case "7":
			c.deleteData(ctx)
		case "8":
			c.syncData(ctx)
		case "9":
			c.healthCheck(ctx)
		case "0":
			fmt.Println("Goodbye!")
			return nil
		default:
			fmt.Println("Invalid option")
		}
	}
}

func (c *CLI) register(ctx context.Context) {
	fmt.Print("Login: ")
	c.scanner.Scan()
	login := c.scanner.Text()

	fmt.Print("Password: ")
	c.scanner.Scan()
	password := c.scanner.Text()

	resp, err := c.client.Register(ctx, login, password)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}

	fmt.Printf("Registration successful! Token: %s\n", resp.Token)
	c.saveToken(resp.Token)
}

func (c *CLI) login(ctx context.Context) {
	fmt.Print("Login: ")
	c.scanner.Scan()
	login := c.scanner.Text()

	fmt.Print("Password: ")
	c.scanner.Scan()
	password := c.scanner.Text()

	resp, err := c.client.Login(ctx, login, password)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	fmt.Printf("Login successful! Token: %s\n", resp.Token)
	c.saveToken(resp.Token)
}

func (c *CLI) saveData(ctx context.Context) {
	fmt.Println("Select data type:")
	fmt.Println("1. Login/Password")
	fmt.Println("2. Text data")
	fmt.Println("3. Binary data")
	fmt.Println("4. Card data")
	fmt.Print("Choice: ")

	c.scanner.Scan()
	typeChoice := c.scanner.Text()

	var dataType models.DataType
	var dataBytes []byte

	switch typeChoice {
	case "1":
		dataType = models.LoginPassword
		fmt.Print("Resource: ")
		c.scanner.Scan()
		resource := c.scanner.Text()

		fmt.Print("Login: ")
		c.scanner.Scan()
		login := c.scanner.Text()

		fmt.Print("Password: ")
		c.scanner.Scan()
		password := c.scanner.Text()

		loginData := models.LoginPasswordData{
			Login:    login,
			Password: password,
			Resource: resource,
		}
		dataBytes, _ = json.Marshal(loginData)
	case "2":
		dataType = models.TextData
		fmt.Print("Enter text: ")
		c.scanner.Scan()
		dataBytes = []byte(c.scanner.Text())
	case "3":
		dataType = models.BinaryData
		fmt.Print("File path: ")
		c.scanner.Scan()
		filePath := c.scanner.Text()

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			return
		}
		dataBytes = content
	case "4":
		dataType = models.CardData
		fmt.Print("Card number: ")
		c.scanner.Scan()
		number := c.scanner.Text()

		fmt.Print("Expiry date: ")
		c.scanner.Scan()
		expiry := c.scanner.Text()

		fmt.Print("CVV: ")
		c.scanner.Scan()
		cvv := c.scanner.Text()

		fmt.Print("Cardholder: ")
		c.scanner.Scan()
		cardholder := c.scanner.Text()

		cardData := models.BankData{
			CardNumber: number,
			ExpiryDate: expiry,
			CVV:        cvv,
			Cardholder: cardholder,
		}
		dataBytes, _ = json.Marshal(cardData)
	default:
		fmt.Println("Invalid data type")
		return
	}

	fmt.Print("Name: ")
	c.scanner.Scan()
	name := c.scanner.Text()

	fmt.Print("Metadata: ")
	c.scanner.Scan()
	metadata := c.scanner.Text()

	data := &models.SecretItemData{
		Type:     dataType,
		Name:     name,
		Metadata: metadata,
		Data:     dataBytes,
	}

	if err := c.client.SaveData(ctx, data); err != nil {
		fmt.Printf("Failed to save data: %v\n", err)
		return
	}

	fmt.Println("Data saved successfully!")
}

func (c *CLI) listData(ctx context.Context) {
	data, err := c.client.ListData(ctx, time.Time{})
	if err != nil {
		fmt.Printf("Failed to list data: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d items:\n", len(data))
	for i, item := range data {
		// Обрезаем длинные данные для красивого отображения
		metadata := item.Metadata
		if len(metadata) > 30 {
			metadata = metadata[:27] + "..."
		}

		fmt.Printf("%d. %s (%s)\n", i+1, item.Name, item.Type)
		fmt.Printf("   Metadata: %s\n", metadata)
		fmt.Printf("   Updated: %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   ID: %s\n", item.ID)
		fmt.Println()
	}
}

func (c *CLI) getData(ctx context.Context) {
	fmt.Println("Get data by:")
	fmt.Println("1. Name")
	fmt.Println("2. ID")
	fmt.Print("Choose option: ")

	c.scanner.Scan()
	choice := c.scanner.Text()

	var data *models.SecretItemData
	var err error

	switch choice {
	case "1":
		fmt.Print("Data Name: ")
		c.scanner.Scan()
		name := c.scanner.Text()
		data, err = c.client.GetDataByName(ctx, name)
	case "2":
		fmt.Print("Data ID: ")
		c.scanner.Scan()
		dataID := c.scanner.Text()
		data, err = c.client.GetData(ctx, dataID)
	default:
		fmt.Println("Invalid option")
		return
	}

	if err != nil {
		fmt.Printf("Failed to get data: %v\n", err)
		return
	}

	c.displayDataDetails(data)
}

func (c *CLI) syncData(ctx context.Context) {

	req := &models.SyncRequest{
		LastSync: time.Now().Add(-24 * time.Hour),
		Data:     []models.SecretItemData{},
	}

	resp, err := c.client.SyncData(ctx, req)
	if err != nil {
		fmt.Printf("Sync failed: %v\n", err)
		return
	}

	fmt.Printf("Sync successful! Received %d items\n", len(resp.Data))
}

func (c *CLI) deleteData(ctx context.Context) {
	fmt.Println("Delete data by:")
	fmt.Println("1. Name")
	fmt.Println("2. ID")
	fmt.Print("Choose option: ")

	c.scanner.Scan()
	choice := c.scanner.Text()

	var err error

	switch choice {
	case "1":
		fmt.Print("Data Name to delete: ")
		c.scanner.Scan()
		name := c.scanner.Text()
		err = c.client.DeleteDataByName(ctx, name)
	case "2":
		fmt.Print("Data ID to delete: ")
		c.scanner.Scan()
		dataID := c.scanner.Text()
		err = c.client.DeleteData(ctx, dataID)
	default:
		fmt.Println("Invalid option")
		return
	}

	if err != nil {
		fmt.Printf("Delete failed: %v\n", err)
		return
	}

	fmt.Println("Data deleted successfully!")
}

func (c *CLI) healthCheck(ctx context.Context) {
	if err := c.client.Health(ctx); err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return
	}
	fmt.Println("Server is healthy!")
}

func (c *CLI) saveToken(token string) {
	tokenFile := filepath.Join(c.config.DataDir, "token")
	if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
		fmt.Printf("Warning: failed to save token: %v\n", err)
	}
}

func LoadToken(dataDir string) string {
	tokenFile := filepath.Join(dataDir, "token")
	content, err := os.ReadFile(tokenFile)
	if err != nil {
		return ""
	}
	return string(content)
}
func (c *CLI) displayDataDetails(data *models.SecretItemData) {
	fmt.Printf("\nData details:\n")
	fmt.Printf("ID: %s\n", data.ID)
	fmt.Printf("Name: %s\n", data.Name)
	fmt.Printf("Type: %s\n", data.Type)
	fmt.Printf("Metadata: %s\n", data.Metadata)
	fmt.Printf("Version: %d\n", data.Version)
	fmt.Printf("Created: %s\n", data.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", data.UpdatedAt.Format(time.RFC3339))

	switch data.Type {
	case models.LoginPassword:
		var loginData models.LoginPasswordData
		if err := json.Unmarshal(data.Data, &loginData); err == nil {
			fmt.Printf("Resource: %s\n", loginData.Resource)
			fmt.Printf("Login: %s\n", loginData.Login)
			fmt.Printf("Password: %s\n", loginData.Password)
		}
	case models.TextData:
		fmt.Printf("Content: %s\n", string(data.Data))
	case models.CardData:
		var cardData models.BankData
		if err := json.Unmarshal(data.Data, &cardData); err == nil {
			fmt.Printf("Card Number: %s\n", cardData.CardNumber)
			fmt.Printf("Expiry: %s\n", cardData.ExpiryDate)
			fmt.Printf("Cardholder: %s\n", cardData.Cardholder)
		}
	case models.BinaryData:
		fmt.Printf("Binary data size: %d bytes\n", len(data.Data))
		fmt.Print("Save to file? (y/n): ")
		c.scanner.Scan()
		if strings.ToLower(c.scanner.Text()) == "y" {
			fmt.Print("File path: ")
			c.scanner.Scan()
			filePath := c.scanner.Text()
			if err := os.WriteFile(filePath, data.Data, 0644); err != nil {
				fmt.Printf("Failed to save file: %v\n", err)
			} else {
				fmt.Printf("Data saved to %s\n", filePath)
			}
		}
	}
}

func (c *CLI) updateData(ctx context.Context) {
	fmt.Print("Data Name to update: ")
	c.scanner.Scan()
	name := c.scanner.Text()

	currentData, err := c.client.GetDataByName(ctx, name)
	if err != nil {
		fmt.Printf("Failed to get data: %v\n", err)
		return
	}

	fmt.Printf("Current data: %s (%s)\n", currentData.Name, currentData.Type)
	fmt.Print("New name (leave empty to keep current): ")
	c.scanner.Scan()
	newName := c.scanner.Text()
	if newName != "" {
		currentData.Name = newName
	}

	fmt.Print("New metadata (leave empty to keep current): ")
	c.scanner.Scan()
	newMetadata := c.scanner.Text()
	if newMetadata != "" {
		currentData.Metadata = newMetadata
	}

	switch currentData.Type {
	case models.TextData:
		fmt.Printf("Current content: %s\n", string(currentData.Data))
		fmt.Print("New content (leave empty to keep current): ")
		c.scanner.Scan()
		newContent := c.scanner.Text()
		if newContent != "" {
			currentData.Data = []byte(newContent)
		}
	case models.LoginPassword:
		var loginData models.LoginPasswordData
		json.Unmarshal(currentData.Data, &loginData)
		fmt.Printf("Current resource: %s\n", loginData.Resource)
		fmt.Print("New resource (leave empty to keep current): ")
		c.scanner.Scan()
		newResource := c.scanner.Text()
		if newResource != "" {
			loginData.Resource = newResource
		}

		fmt.Printf("Current login: %s\n", loginData.Login)
		fmt.Print("New login (leave empty to keep current): ")
		c.scanner.Scan()
		newLogin := c.scanner.Text()
		if newLogin != "" {
			loginData.Login = newLogin
		}

		fmt.Print("New password (leave empty to keep current): ")
		c.scanner.Scan()
		newPassword := c.scanner.Text()
		if newPassword != "" {
			loginData.Password = newPassword
		}

		currentData.Data, _ = json.Marshal(loginData)
	}

	if err := c.client.UpdateData(ctx, currentData); err != nil {
		fmt.Printf("Failed to update data: %v\n", err)
		return
	}

	fmt.Println("Data updated successfully!")
}

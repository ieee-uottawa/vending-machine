package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stianeikeland/go-rpio"
)

// VendingMachine represents our main application with all dependencies
// Using a struct allows us to encapsulate state and makes testing easier
type VendingMachine struct {
	relayPins       map[int]rpio.Pin    // Maps logical pin numbers to GPIO pins
	slotRelays      map[string][]int    // Maps slot names to relay pin numbers
	processedOrders map[string]struct{} // Set of processed order IDs
	ordersMutex     sync.RWMutex        // Protects processedOrders from race conditions
	httpClient      *http.Client        // Reusable HTTP client for better performance
	squareToken     string              // Square API access token
	squareAPIBase   string              // Square API base URL
}

// SquareWebhookPayload represents the incoming webhook from Square
// Using struct tags for JSON marshaling
type SquareWebhookPayload struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Data struct {
		Object struct {
			Payment struct {
				Status  string `json:"status"`
				OrderID string `json:"order_id"`
			} `json:"payment"`
		} `json:"object"`
	} `json:"data"`
}

// SquareOrderResponse represents the order data from Square API
type SquareOrderResponse struct {
	Order struct {
		LineItems []struct {
			CatalogObjectID string `json:"catalog_object_id"`
			UID             string `json:"uid"`
		} `json:"line_items"`
	} `json:"order"`
}

// SquareCatalogResponse represents catalog object data from Square API
type SquareCatalogResponse struct {
	Object struct {
		CustomAttributeValues map[string]struct {
			CustomAttributeDefinitionID string   `json:"custom_attribute_definition_id"`
			SelectionUIDValues          []string `json:"selection_uid_values"`
		} `json:"custom_attribute_values"`
	} `json:"object"`
}

// SquareDefinitionResponse represents custom attribute definition from Square API
type SquareDefinitionResponse struct {
	Object struct {
		CustomAttributeDefinitionData struct {
			SelectionConfig struct {
				AllowedSelections []struct {
					UID  string `json:"uid"`
					Name string `json:"name"`
				} `json:"allowed_selections"`
			} `json:"selection_config"`
		} `json:"custom_attribute_definition_data"`
	} `json:"object"`
}

// NewVendingMachine creates a new vending machine instance
func NewVendingMachine() *VendingMachine {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Define relay pin mapping (logical pin -> physical GPIO pin)
	relayPins := map[int]rpio.Pin{
		1: 2, 2: 3, 3: 4, 4: 17, 5: 27, 6: 22, 7: 10, 8: 9,
		9: 11, 10: 5, 11: 6, 12: 13, 13: 19, 14: 26, 15: 14, 16: 15,
	}

	// Define slot to relay mapping
	slotRelays := map[string][]int{
		"A1": {3, 12, 13, 14}, "A2": {3, 7, 13, 14}, "A3": {3, 7, 12, 14}, "A4": {3, 7, 12, 13},
		"B1": {2, 12, 13, 14}, "B2": {2, 7, 13, 14}, "B3": {2, 7, 12, 14}, "B4": {2, 7, 12, 13},
		"C1": {5, 12, 13, 14}, "C2": {5, 7, 13, 14}, "C3": {5, 7, 12, 14}, "C4": {5, 7, 12, 13},
		"D1": {4, 16, 15, 14, 13, 12, 10, 8}, "D2": {4, 16, 15, 14, 13, 10, 8, 7},
		"D3": {4, 16, 15, 14, 12, 10, 8, 7}, "D4": {4, 16, 15, 13, 12, 10, 8, 7},
		"D5": {4, 16, 14, 13, 12, 7, 8, 10}, "D6": {4, 16, 14, 13, 12, 7, 8, 15},
		"D7": {4, 15, 14, 13, 12, 10, 8, 7}, "D8": {4, 16, 15, 14, 13, 12, 10, 7},
		"E1": {1, 16, 15, 14, 13, 12, 10, 8}, "E2": {1, 16, 15, 14, 13, 10, 8, 7},
		"E3": {1, 16, 15, 14, 12, 10, 8, 7}, "E4": {1, 16, 15, 13, 12, 10, 8, 7},
		"E5": {1, 16, 14, 13, 12, 7, 8, 10}, "E6": {1, 16, 14, 13, 12, 7, 8, 15},
		"E7": {1, 15, 14, 13, 12, 10, 8, 7}, "E8": {1, 16, 15, 14, 13, 12, 10, 7},
		"F1": {6, 12, 13, 14}, "F2": {6, 7, 13, 14}, "F3": {6, 7, 12, 14}, "F4": {6, 7, 12, 13},
	}

	return &VendingMachine{
		relayPins:       relayPins,
		slotRelays:      slotRelays,
		processedOrders: make(map[string]struct{}),
		// Using a timeout for HTTP client - prevents hanging requests
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		//squareToken: os.Getenv("SQUARE_ACCESS_TOKEN_SANDBOX"),
		//squareAPIBase: "https://connect.squareupsandbox.com/v2",
		squareToken:   os.Getenv("SQUARE_ACCESS_TOKEN_PROD"),
		squareAPIBase: "https://connect.squareup.com/v2",
	}
}

// InitializeGPIO sets up the GPIO pins for relay control
func (vm *VendingMachine) InitializeGPIO() error {
	log.Println("Initializing GPIO...")

	if err := rpio.Open(); err != nil {
		return fmt.Errorf("failed to open GPIO: %w", err)
	}

	// Initialize all relay pins
	for logicalPin, physicalPin := range map[int]int{
		1: 2, 2: 3, 3: 4, 4: 17, 5: 27, 6: 22, 7: 10, 8: 9,
		9: 11, 10: 5, 11: 6, 12: 13, 13: 19, 14: 26, 15: 14, 16: 15,
	} {
		pin := rpio.Pin(physicalPin)
		pin.Output()
		pin.High() // Relays are active LOW, so HIGH = off
		vm.relayPins[logicalPin] = pin
	}

	log.Println("GPIO initialization complete")
	return nil
}

// DispenseItem activates the relays for a specific slot
// Using a goroutine here for non-blocking operation (like Python's background tasks)
func (vm *VendingMachine) DispenseItem(slotLabel string) {
	go func() {
		relays, exists := vm.slotRelays[slotLabel]
		if !exists {
			log.Printf("Unknown slot label: %s", slotLabel)
			return
		}

		log.Printf("Dispensing item from slot %s", slotLabel)

		// Activate relays (set LOW)
		for _, logicalPin := range relays {
			if pin, exists := vm.relayPins[logicalPin]; exists {
				pin.Low()
			} else {
				log.Printf("Warning: Pin %d not found", logicalPin)
			}
		}

		// Wait for dispense duration
		time.Sleep(3300 * time.Millisecond)

		// Deactivate relays (set HIGH)
		for _, logicalPin := range relays {
			if pin, exists := vm.relayPins[logicalPin]; exists {
				pin.High()
			}
		}

		log.Printf("Finished dispensing from slot %s", slotLabel)
	}()
}

// makeSquareAPIRequest is a helper function to make authenticated requests to Square API
func (vm *VendingMachine) makeSquareAPIRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers for Square API
	req.Header.Set("Authorization", "Bearer "+vm.squareToken)
	req.Header.Set("Square-Version", "2025-07-16")
	req.Header.Set("Content-Type", "application/json")

	return vm.httpClient.Do(req)
}

// HandleSquareEvent processes Square webhook events
func (vm *VendingMachine) HandleSquareEvent(ctx context.Context, payload SquareWebhookPayload) error {
	// Check if this is a payment completion event
	if payload.Type != "payment.updated" ||
		payload.Data.Object.Payment.Status != "COMPLETED" ||
		payload.Data.Object.Payment.OrderID == "" {
		return fmt.Errorf("ignoring non-payment event")
	}

	orderID := payload.Data.Object.Payment.OrderID

	// Thread-safe check for duplicate orders using mutex
	vm.ordersMutex.Lock()
	if _, processed := vm.processedOrders[orderID]; processed {
		vm.ordersMutex.Unlock()
		log.Printf("Ignoring duplicate webhook for order %s", orderID)
		return fmt.Errorf("order already processed")
	}
	vm.processedOrders[orderID] = struct{}{} // struct{}{} is memory-efficient empty value
	vm.ordersMutex.Unlock()

	log.Printf("Processing order %s", orderID)

	// Fetch order details from Square API
	orderURL := fmt.Sprintf("%s/orders/%s", vm.squareAPIBase, orderID)
	resp, err := vm.makeSquareAPIRequest(ctx, orderURL)
	if err != nil {
		return fmt.Errorf("failed to fetch order: %w", err)
	}
	defer resp.Body.Close() // Always close response bodies to prevent memory leaks

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error fetching order %s: %s", orderID, string(body))
	}

	var orderData SquareOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderData); err != nil {
		return fmt.Errorf("failed to decode order response: %w", err)
	}

	// Process each line item in the order
	for _, item := range orderData.Order.LineItems {
		catalogObjectID := item.CatalogObjectID
		if catalogObjectID == "" {
			catalogObjectID = item.UID
		}
		if catalogObjectID == "" {
			log.Printf("No catalog object ID found for item")
			continue
		}

		slotLabel, err := vm.getSlotLabelFromCatalogObject(ctx, catalogObjectID)
		if err != nil {
			log.Printf("Error getting slot label: %v", err)
			continue
		}

		if slotLabel != "" {
			vm.DispenseItem(slotLabel)
		}
	}

	return nil
}

// getSlotLabelFromCatalogObject fetches the slot label from Square's catalog system
func (vm *VendingMachine) getSlotLabelFromCatalogObject(ctx context.Context, catalogObjectID string) (string, error) {
	// Fetch catalog object
	objURL := fmt.Sprintf("%s/catalog/object/%s", vm.squareAPIBase, catalogObjectID)
	resp, err := vm.makeSquareAPIRequest(ctx, objURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch catalog object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching catalog object: status %d", resp.StatusCode)
	}

	var objData SquareCatalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&objData); err != nil {
		return "", fmt.Errorf("failed to decode catalog response: %w", err)
	}

	// Extract custom attributes
	if len(objData.Object.CustomAttributeValues) == 0 {
		return "", fmt.Errorf("no custom attributes found")
	}

	// Get the first custom attribute (assuming only one exists)
	var firstAttr struct {
		CustomAttributeDefinitionID string   `json:"custom_attribute_definition_id"`
		SelectionUIDValues          []string `json:"selection_uid_values"`
	}

	for _, attr := range objData.Object.CustomAttributeValues {
		firstAttr = attr
		break
	}

	if len(firstAttr.SelectionUIDValues) == 0 {
		return "", fmt.Errorf("no selection UID values found")
	}

	selectionUID := firstAttr.SelectionUIDValues[0]
	definitionID := firstAttr.CustomAttributeDefinitionID

	// Fetch custom attribute definition
	defURL := fmt.Sprintf("%s/catalog/object/%s", vm.squareAPIBase, definitionID)
	defResp, err := vm.makeSquareAPIRequest(ctx, defURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch definition: %w", err)
	}
	defer defResp.Body.Close()

	if defResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error fetching definition: status %d", defResp.StatusCode)
	}

	var defData SquareDefinitionResponse
	if err := json.NewDecoder(defResp.Body).Decode(&defData); err != nil {
		return "", fmt.Errorf("failed to decode definition response: %w", err)
	}

	// Find the slot label for this selection UID
	for _, selection := range defData.Object.CustomAttributeDefinitionData.SelectionConfig.AllowedSelections {
		if selection.UID == selectionUID {
			return selection.Name, nil
		}
	}

	return "", fmt.Errorf("selection UID %s not found in definition", selectionUID)
}

// setupRoutes configures the HTTP routes using Gin
func (vm *VendingMachine) setupRoutes() *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	// Square webhook endpoint
	// Using a closure to capture vm instance - this is how we access methods in handlers
	router.POST("/webhook/square", func(c *gin.Context) {
		var payload SquareWebhookPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			log.Printf("Invalid webhook payload: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		// Process webhook in background (like Python's BackgroundTasks)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := vm.HandleSquareEvent(ctx, payload); err != nil {
				log.Printf("Error processing webhook: %v", err)
			}
		}()

		c.JSON(http.StatusOK, gin.H{
			"message": "Webhook received and processing started",
		})
	})

	return router
}

func main() {
	// Create vending machine instance
	vm := NewVendingMachine()

	// Initialize GPIO
	if err := vm.InitializeGPIO(); err != nil {
		log.Fatalf("Failed to initialize GPIO: %v", err)
	}
	// Defer cleanup - Go's way of ensuring cleanup happens
	defer rpio.Close()

	// Setup HTTP routes
	router := vm.setupRoutes()

	// Start server
	log.Println("Starting server on :8000")
	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

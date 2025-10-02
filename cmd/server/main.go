package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	square "github.com/square/square-go-sdk"
	"github.com/square/square-go-sdk/client"
	"github.com/square/square-go-sdk/option"
	"github.com/stianeikeland/go-rpio"
)

// VendingMachine represents our main application with all dependencies
// Using a struct allows us to encapsulate state and makes testing easier
type VendingMachine struct {
	relayPins       map[int]rpio.Pin    // Maps logical pin numbers to GPIO pins
	slotRelays      map[string][]int    // Maps slot names to relay pin numbers
	processedOrders map[string]struct{} // Set of processed order IDs
	ordersMutex     sync.RWMutex        // Protects processedOrders from race conditions
	squareClient    *client.Client      // Square SDK client
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
		// Create Square SDK client with production token
		squareClient: client.NewClient(
			//option.WithToken(os.Getenv("SQUARE_ACCESS_TOKEN_SANDBOX")),
			//option.WithBaseURL("https://connect.squareupsandbox.com/v2"),
			option.WithToken(os.Getenv("SQUARE_ACCESS_TOKEN_PROD")),
			option.WithBaseURL("https://connect.squareup.com"),
		),
	}
}

// InitializeGPIO sets up the GPIO pins for relay control
func (vm *VendingMachine) InitializeGPIO() error {
	log.Println("Initializing GPIO...")

	if err := rpio.Open(); err != nil {
		return fmt.Errorf("failed to open GPIO (are you on the pi?): %w", err)
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

	// Fetch order details using Square SDK
	getOrderRequest := &square.GetOrdersRequest{
		OrderID: orderID,
	}

	orderResponse, err := vm.squareClient.Orders.Get(ctx, getOrderRequest)
	if err != nil {
		return fmt.Errorf("failed to fetch order: %w", err)
	}

	if orderResponse.Order == nil {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Process each line item in the order
	for _, item := range orderResponse.Order.LineItems {
		catalogObjectID := ""
		if item.CatalogObjectID != nil {
			catalogObjectID = *item.CatalogObjectID
		} else if item.UID != nil {
			catalogObjectID = *item.UID
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

// getSlotLabelFromCatalogObject fetches the slot label from Square's catalog system using SDK
func (vm *VendingMachine) getSlotLabelFromCatalogObject(ctx context.Context, catalogObjectID string) (string, error) {
	// Fetch catalog object using Square SDK
	batchGetRequest := &square.BatchGetCatalogObjectsRequest{
		ObjectIDs: []string{catalogObjectID},
	}

	catalogResponse, err := vm.squareClient.Catalog.BatchGet(ctx, batchGetRequest)
	if err != nil {
		return "", fmt.Errorf("failed to fetch catalog object: %w", err)
	}

	if len(catalogResponse.Objects) == 0 {
		return "", fmt.Errorf("catalog object not found: %s", catalogObjectID)
	}

	catalogObject := catalogResponse.Objects[0]

	// Extract custom attributes from the appropriate object type
	var customAttributeValues map[string]*square.CatalogCustomAttributeValue

	if catalogObject.Item != nil {
		customAttributeValues = catalogObject.Item.CustomAttributeValues
	} else if catalogObject.ItemVariation != nil {
		customAttributeValues = catalogObject.ItemVariation.CustomAttributeValues
	}

	// Check if we found custom attributes
	if len(customAttributeValues) == 0 {
		return "", fmt.Errorf("no custom attributes found")
	}

	// Get the first custom attribute (assuming only one exists)
	var firstAttrDefinitionID string
	var firstAttrSelectionUIDs []string

	for _, attr := range customAttributeValues {
		if attr.CustomAttributeDefinitionID != nil {
			firstAttrDefinitionID = *attr.CustomAttributeDefinitionID
		}
		if attr.SelectionUIDValues != nil {
			firstAttrSelectionUIDs = attr.SelectionUIDValues
		}
		break
	}

	if len(firstAttrSelectionUIDs) == 0 {
		return "", fmt.Errorf("no selection UID values found")
	}

	selectionUID := firstAttrSelectionUIDs[0]

	// Fetch custom attribute definition using Square SDK
	defBatchGetRequest := &square.BatchGetCatalogObjectsRequest{
		ObjectIDs: []string{firstAttrDefinitionID},
	}

	defResponse, err := vm.squareClient.Catalog.BatchGet(ctx, defBatchGetRequest)
	if err != nil {
		return "", fmt.Errorf("failed to fetch definition: %w", err)
	}

	if len(defResponse.Objects) == 0 {
		return "", fmt.Errorf("definition object not found: %s", firstAttrDefinitionID)
	}

	definitionObject := defResponse.Objects[0]

	// Find the slot label for this selection UID
	if definitionObject.CustomAttributeDefinition != nil &&
		definitionObject.CustomAttributeDefinition.CustomAttributeDefinitionData != nil &&
		definitionObject.CustomAttributeDefinition.CustomAttributeDefinitionData.SelectionConfig != nil &&
		definitionObject.CustomAttributeDefinition.CustomAttributeDefinitionData.SelectionConfig.AllowedSelections != nil {

		for _, selection := range definitionObject.CustomAttributeDefinition.CustomAttributeDefinitionData.SelectionConfig.AllowedSelections {
			if selection.UID != nil && selection.Name != "" && *selection.UID == selectionUID {
				return selection.Name, nil
			}
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


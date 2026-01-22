package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Pet struct demonstrates struct types with various field types
type Pet struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Species      string            `json:"species"`
	Breed        string            `json:"breed"`
	Age          int               `json:"age"`
	Gender       string            `json:"gender"`
	Description  string            `json:"description"`
	Status       string            `json:"status"` // Available, Adopted, Under Care
	IsVaccinated bool              `json:"isVaccinated"`
	CreatedAt    time.Time         `json:"createdAt"`
	Tags         []string          `json:"tags"`       // Array/Slice
	Attributes   map[string]string `json:"attributes"` // Map
}

// Service struct for service offerings
type Service struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Duration    int      `json:"duration"` // in minutes
	Available   bool     `json:"available"`
	Features    []string `json:"features"`
}

// ContactForm struct for contact submissions
type ContactForm struct {
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Purpose string    `json:"purpose"`
	Message string    `json:"message"`
	SentAt  time.Time `json:"sentAt"`
}

// ServiceBooking struct for service bookings
type ServiceBooking struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"serviceId"`
	PetName   string    `json:"petName"`
	OwnerName string    `json:"ownerName"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Date      string    `json:"date"`
	Time      string    `json:"time"`
	Notes     string    `json:"notes"`
	Status    string    `json:"status"`
	BookedAt  time.Time `json:"bookedAt"`
}

// Global variables demonstrating various types
var (
	// 1. VARIABLES, VALUES AND TYPES
	// Basic types
	serverStartTime time.Time = time.Now()
	requestCount    int       = 0
	serverVersion   string    = "1.0.0"
	isProduction    bool      = false
	maxPets         int       = 100
	defaultTimeout  float64   = 30.0

	petCategories   [4]string = [4]string{"Dogs", "Cats", "Birds", "Others"}
	priorities      [3]int    = [3]int{1, 2, 3}
	pets            []Pet
	services        []Service
	contactMessages []ContactForm
	bookings        []ServiceBooking
	allowedOrigins  []string

	petsByID     map[string]*Pet
	servicesByID map[string]*Service
	bookingsByID map[string]*ServiceBooking
	statusCounts map[string]int
	serviceStats map[string]map[string]interface{}
)

// Initialize data function
func initializeData() {
	// Initialize maps
	petsByID = make(map[string]*Pet)
	servicesByID = make(map[string]*Service)
	bookingsByID = make(map[string]*ServiceBooking)
	statusCounts = make(map[string]int)
	serviceStats = make(map[string]map[string]interface{})

	// Initialize slices
	pets = make([]Pet, 0, maxPets)
	services = make([]Service, 0, 20)
	contactMessages = make([]ContactForm, 0)
	bookings = make([]ServiceBooking, 0)
	allowedOrigins = []string{"http://localhost:8080", "http://127.0.0.1:8080"}

	// Sample pets data
	samplePets := []Pet{
		{
			ID:           "pet-001",
			Name:         "Max",
			Species:      "Dog",
			Breed:        "Golden Retriever",
			Age:          3,
			Gender:       "Male",
			Description:  "Friendly and energetic dog looking for a loving home",
			Status:       "Available",
			IsVaccinated: true,
			CreatedAt:    time.Now().AddDate(0, -2, 0),
			Tags:         []string{"Friendly", "Energetic", "House-trained"},
			Attributes:   map[string]string{"Color": "Golden", "Size": "Large", "Weight": "30kg"},
		},
		{
			ID:           "pet-002",
			Name:         "Luna",
			Species:      "Cat",
			Breed:        "Persian",
			Age:          2,
			Gender:       "Female",
			Description:  "Calm and affectionate cat, perfect for apartments",
			Status:       "Available",
			IsVaccinated: true,
			CreatedAt:    time.Now().AddDate(0, -1, -15),
			Tags:         []string{"Calm", "Indoor", "Affectionate"},
			Attributes:   map[string]string{"Color": "White", "Size": "Medium", "Weight": "4kg"},
		},
		{
			ID:           "pet-003",
			Name:         "Charlie",
			Species:      "Dog",
			Breed:        "Beagle",
			Age:          1,
			Gender:       "Male",
			Description:  "Playful puppy with lots of energy",
			Status:       "Under Care",
			IsVaccinated: false,
			CreatedAt:    time.Now().AddDate(0, 0, -10),
			Tags:         []string{"Playful", "Young", "Needs Training"},
			Attributes:   map[string]string{"Color": "Brown and White", "Size": "Medium", "Weight": "12kg"},
		},
	}

	// 2. LOOPING STRUCTURES - for loop with range
	for i, pet := range samplePets {
		pets = append(pets, pet)
		petsByID[pet.ID] = &pets[i]
		statusCounts[pet.Status]++
	}

	// Sample services data
	sampleServices := []Service{
		{
			ID:          "svc-001",
			Name:        "Pet Grooming",
			Category:    "Care",
			Description: "Complete grooming service including bath, haircut, and nail trimming",
			Price:       1500.00,
			Duration:    90,
			Available:   true,
			Features:    []string{"Bath", "Haircut", "Nail Trimming", "Ear Cleaning"},
		},
		{
			ID:          "svc-002",
			Name:        "Veterinary Checkup",
			Category:    "Medical",
			Description: "Comprehensive health checkup by certified veterinarians",
			Price:       2000.00,
			Duration:    60,
			Available:   true,
			Features:    []string{"Physical Exam", "Vaccination", "Health Certificate"},
		},
		{
			ID:          "svc-003",
			Name:        "Dog Training",
			Category:    "Training",
			Description: "Basic obedience training for dogs",
			Price:       3000.00,
			Duration:    120,
			Available:   true,
			Features:    []string{"Basic Commands", "Socialization", "Behavioral Correction"},
		},
		{
			ID:          "svc-004",
			Name:        "Pet Boarding",
			Category:    "Care",
			Description: "Safe and comfortable boarding facilities",
			Price:       800.00,
			Duration:    1440, // per day
			Available:   true,
			Features:    []string{"24/7 Care", "Play Area", "Regular Meals"},
		},
	}

	// 2. LOOPING STRUCTURES - traditional for loop
	for i := 0; i < len(sampleServices); i++ {
		services = append(services, sampleServices[i])
		servicesByID[sampleServices[i].ID] = &services[i]

		// Initialize service stats
		serviceStats[sampleServices[i].ID] = map[string]interface{}{
			"bookings":  0,
			"revenue":   0.0,
			"rating":    4.5,
			"available": sampleServices[i].Available,
		}
	}
}

// Helper function to demonstrate control flow
func validatePet(pet Pet) (bool, []string) {
	// 2. CONTROL FLOW - if-else statements
	errors := make([]string, 0)

	if pet.Name == "" {
		errors = append(errors, "Pet name is required")
	}

	if pet.Species == "" {
		errors = append(errors, "Species is required")
	}

	if pet.Age < 0 || pet.Age > 30 {
		errors = append(errors, "Age must be between 0 and 30")
	}

	// 2. CONTROL FLOW - switch statement
	switch pet.Status {
	case "Available", "Adopted", "Under Care":
		// Valid status
	default:
		errors = append(errors, "Invalid status")
	}

	return len(errors) == 0, errors
}

// Helper function to calculate statistics
func calculateStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count pets by status
	stats["petsByStatus"] = statusCounts

	// Count pets by species
	speciesCount := make(map[string]int)
	for _, pet := range pets {
		speciesCount[pet.Species]++
	}
	stats["petsBySpecies"] = speciesCount

	// Calculate averages
	if len(pets) > 0 {
		totalAge := 0
		vaccinatedCount := 0

		// 2. LOOPING - for range over slice
		for _, pet := range pets {
			totalAge += pet.Age
			if pet.IsVaccinated {
				vaccinatedCount++
			}
		}

		stats["averageAge"] = float64(totalAge) / float64(len(pets))
		stats["vaccinationRate"] = float64(vaccinatedCount) / float64(len(pets)) * 100
	}

	stats["totalPets"] = len(pets)
	stats["totalServices"] = len(services)
	stats["totalBookings"] = len(bookings)
	stats["totalMessages"] = len(contactMessages)

	return stats
}

// Filter pets by criteria
func filterPets(species, status string, minAge, maxAge int) []Pet {
	filtered := make([]Pet, 0)

	// 2. LOOPING AND CONTROL FLOW combined
	for _, pet := range pets {
		include := true

		if species != "" && pet.Species != species {
			include = false
		}

		if status != "" && pet.Status != status {
			include = false
		}

		if minAge > 0 && pet.Age < minAge {
			include = false
		}

		if maxAge > 0 && pet.Age > maxAge {
			include = false
		}

		if include {
			filtered = append(filtered, pet)
		}
	}

	return filtered
}

// HTTP Handlers

// CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// GET /api/pets - Get all pets with optional filtering
func getPetsHandler(w http.ResponseWriter, r *http.Request) {
	requestCount++ // Increment counter

	query := r.URL.Query()
	species := query.Get("species")
	status := query.Get("status")

	var result []Pet

	// 2. CONTROL FLOW - if-else
	if species == "" && status == "" {
		result = pets
	} else {
		result = filterPets(species, status, 0, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(result),
		"data":    result,
	})
}

// GET /api/pets/:id - Get single pet
func getPetByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/pets/")
	petID := path

	pet, exists := petsByID[petID]

	w.Header().Set("Content-Type", "application/json")

	// 2. CONTROL FLOW - if-else
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Pet not found",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    pet,
	})
}

// POST /api/pets - Add new pet
func addPetHandler(w http.ResponseWriter, r *http.Request) {
	var newPet Pet

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&newPet); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid JSON data",
		})
		return
	}

	// Validate pet
	valid, errors := validatePet(newPet)
	if !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Validation failed",
			"errors":  errors,
		})
		return
	}

	// Set defaults
	newPet.ID = fmt.Sprintf("pet-%03d", len(pets)+1)
	newPet.CreatedAt = time.Now()

	pets = append(pets, newPet)
	petsByID[newPet.ID] = &pets[len(pets)-1]
	statusCounts[newPet.Status]++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Pet added successfully",
		"data":    newPet,
	})
}

// GET /api/services - Get all services
func getServicesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	category := query.Get("category")

	var result []Service

	// 2. CONTROL FLOW and LOOPING
	if category == "" {
		result = services
	} else {
		// Filter by category
		for _, service := range services {
			if service.Category == category {
				result = append(result, service)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(result),
		"data":    result,
	})
}

// POST /api/bookings - Create booking
func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	var booking ServiceBooking

	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid JSON data",
		})
		return
	}

	// Set defaults
	booking.ID = fmt.Sprintf("book-%03d", len(bookings)+1)
	booking.BookedAt = time.Now()
	booking.Status = "Pending"

	// Add to collections
	bookings = append(bookings, booking)
	bookingsByID[booking.ID] = &bookings[len(bookings)-1]

	// Update service stats
	if stats, exists := serviceStats[booking.ServiceID]; exists {
		stats["bookings"] = stats["bookings"].(int) + 1
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Booking created successfully",
		"data":    booking,
	})
}

// POST /api/contact - Submit contact form
func submitContactHandler(w http.ResponseWriter, r *http.Request) {
	var contact ContactForm

	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid JSON data",
		})
		return
	}

	contact.SentAt = time.Now()
	contactMessages = append(contactMessages, contact)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Message sent successfully",
	})
}

// GET /api/statistics - Get system statistics
func getStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	stats := calculateStatistics()

	// Add runtime stats
	stats["serverVersion"] = serverVersion
	stats["uptime"] = time.Since(serverStartTime).String()
	stats["requestCount"] = requestCount
	stats["isProduction"] = isProduction

	// 4. MAP - accessing nested maps
	stats["serviceStats"] = serviceStats

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

func main() {
	// Initialize data structures
	initializeData()

	// Get MongoDB connection string from environment variable
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("⚠ MONGODB_URI not set, running without database")
	} else {
		// Create a context with timeout for connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Connect to MongoDB
		log.Println("Connecting to MongoDB...")
		client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
		if err != nil {
			log.Printf("Failed to connect to MongoDB: %v", err)
		} else {
			defer func() {
				if err := client.Disconnect(context.Background()); err != nil {
					log.Println("Error disconnecting from MongoDB:", err)
				}
			}()

			// Ping MongoDB to verify connection
			if err := client.Ping(ctx, nil); err != nil {
				log.Printf("Failed to ping MongoDB: %v", err)
			} else {
				log.Println("✓ Successfully connected to MongoDB!")
			}
		}
	}

	// Static file routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/service.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "service.html")
	})

	// API routes demonstrating Go concepts
	http.HandleFunc("/api/pets", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		// 2. CONTROL FLOW - switch on HTTP method
		switch r.Method {
		case "GET":
			// Check if it's a single pet request
			if len(strings.TrimPrefix(r.URL.Path, "/api/pets/")) > 0 && r.URL.Path != "/api/pets" && r.URL.Path != "/api/pets/" {
				getPetByIDHandler(w, r)
			} else {
				getPetsHandler(w, r)
			}
		case "POST":
			addPetHandler(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Method not allowed",
			})
		}
	}))

	http.HandleFunc("/api/pets/", enableCORS(getPetByIDHandler))
	http.HandleFunc("/api/services", enableCORS(getServicesHandler))
	http.HandleFunc("/api/bookings", enableCORS(createBookingHandler))
	http.HandleFunc("/api/contact", enableCORS(submitContactHandler))
	http.HandleFunc("/api/statistics", enableCORS(getStatisticsHandler))

	// Log initialization info
	log.Println("==============================================")
	log.Println("🐾 Pawtner Hope Foundation Server")
	log.Println("==============================================")
	log.Printf("Version: %s\n", serverVersion)
	log.Printf("Initialized with %d pets\n", len(pets))
	log.Printf("Initialized with %d services\n", len(services))
	log.Println("==============================================")
	log.Println("API Endpoints:")
	log.Println("  GET    /api/pets          - Get all pets")
	log.Println("  GET    /api/pets/:id      - Get pet by ID")
	log.Println("  POST   /api/pets          - Add new pet")
	log.Println("  GET    /api/services      - Get all services")
	log.Println("  POST   /api/bookings      - Create booking")
	log.Println("  POST   /api/contact       - Submit contact form")
	log.Println("  GET    /api/statistics    - Get statistics")
	log.Println("==============================================")
	log.Println("Server starting on http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

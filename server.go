package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// 5. FUNCTIONS AND ERROR HANDLING
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrTokenExpired       = errors.New("token has expired")
	ErrPetNotFound        = errors.New("pet not found")
	ErrInvalidPayment     = errors.New("invalid payment details")
	ErrEmailFailed        = errors.New("email delivery failed")
)

// 6. INTERFACE
type Filterable interface {
	Filter(pets []Pet) []Pet
	Name() string
}

// 4. MAP AND STRUCTS

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
	Tags         []string          `json:"tags"`       // 3. ARRAY AND SLICE
	Attributes   map[string]string `json:"attributes"` // 4. MAP AND STRUCTS
}

type Medical struct {
	Vaccinated  bool     `json:"vaccinated"`
	Medications []string `json:"medications"`
	LastCheckup string   `json:"lastCheckup"`
	HealthNotes string   `json:"healthNotes"`
}

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

type ContactForm struct {
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Purpose string    `json:"purpose"`
	Message string    `json:"message"`
	SentAt  time.Time `json:"sentAt"`
}

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

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // excluded from JSON output
	Role      string    `json:"role"`
	IsAdmin   bool      `json:"isAdmin" bson:"isadmin"`
	CreatedAt time.Time `json:"createdAt"`
	IsActive  bool      `json:"isActive"`
}

type AuthToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	Role      string    `json:"role"`
	IsAdmin   bool      `json:"isAdmin"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
}

type Donation struct {
	ID                 string    `json:"id"`
	DonorName          string    `json:"donorName"`
	DonorEmail         string    `json:"donorEmail"`
	Amount             float64   `json:"amount"`
	PaymentMethod      string    `json:"paymentMethod"`
	TransactionID      string    `json:"transactionId"`
	Status             string    `json:"status"` // Pending, Completed, Failed
	CreatedAt          time.Time `json:"createdAt"`
	PaymentViaDeeplink bool      `json:"paymentViaDeeplink"` // true when paid via mobile UPI deeplink
}

type Receipt struct {
	ReceiptID  string    `json:"receiptId"`
	DonationID string    `json:"donationId"`
	DonorName  string    `json:"donorName"`
	Amount     float64   `json:"amount"`
	IssuedAt   time.Time `json:"issuedAt"`
	Message    string    `json:"message"`
}

type AdoptionInquiry struct {
	ID          string    `json:"id"`
	PetID       string    `json:"petId"`
	AdopterName string    `json:"adopterName"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Message     string    `json:"message"`
	Status      string    `json:"status"` // Pending, Approved, Rejected
	CreatedAt   time.Time `json:"createdAt"`
}

// 11. GOROUTINES AND CHANNELS
type NotificationJob struct {
	To      string
	Subject string
	Body    string
	JobType string
}

type PaymentConfirmation struct {
	DonationID    string
	Success       bool
	TransactionID string
	Error         string
}

type PendingRegistration struct {
	Email          string
	Username       string
	HashedPassword string
	Code           string
	ExpiresAt      time.Time
}

// SMTP config (loaded from .env)
var (
	smtpUser string
	smtpPass string
	smtpHost string = "smtp.gmail.com"
	smtpPort string = "587"
)

// loadEnv reads KEY=VALUE lines from a .env file and calls os.Setenv.
func loadEnv(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return // .env is optional
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" { // don't override real env vars
				os.Setenv(key, val)
			}
		}
	}
}

// 1. VARIABLES, VALUES AND TYPES
var (
	serverStartTime time.Time = time.Now()
	requestCount    int       = 0
	serverVersion   string    = "1.0.0"
	isProduction    bool      = false
	maxPets         int       = 100
	defaultTimeout  float64   = 30.0

	// 3. ARRAY AND SLICE
	petCategories [4]string = [4]string{"Dogs", "Cats", "Birds", "Others"}
	priorities    [3]int    = [3]int{1, 2, 3}

	pets            []Pet
	services        []Service
	contactMessages []ContactForm
	bookings        []ServiceBooking
	users           []User
	donations       []Donation
	inquiries       []AdoptionInquiry
	allowedOrigins  []string

	// 4. MAP AND STRUCTS
	petsByID     map[string]*Pet
	servicesByID map[string]*Service
	bookingsByID map[string]*ServiceBooking
	usersByEmail map[string]*User
	tokenStore   map[string]*AuthToken
	statusCounts map[string]int
	serviceStats map[string]map[string]interface{}
	petsByBreed  map[string][]string

	// 10. CONCURRENCY
	notificationCh   chan NotificationJob
	paymentCh        chan Donation
	paymentConfirmCh chan PaymentConfirmation
	mu               sync.Mutex

	emailShouldFail bool = false

	// MongoDB
	mongoClient *mongo.Client
	mongoDB     *mongo.Database

	// Pending email verifications
	pendingRegs map[string]*PendingRegistration
)

func initializeData() {
	petsByID = make(map[string]*Pet)
	servicesByID = make(map[string]*Service)
	bookingsByID = make(map[string]*ServiceBooking)
	usersByEmail = make(map[string]*User)
	tokenStore = make(map[string]*AuthToken)
	statusCounts = make(map[string]int)
	serviceStats = make(map[string]map[string]interface{})
	petsByBreed = make(map[string][]string)

	// 3. ARRAY AND SLICE
	pets = make([]Pet, 0, maxPets)
	services = make([]Service, 0, 20)
	contactMessages = make([]ContactForm, 0)
	bookings = make([]ServiceBooking, 0)
	users = make([]User, 0)
	donations = make([]Donation, 0)
	inquiries = make([]AdoptionInquiry, 0)
	allowedOrigins = []string{"http://localhost:8080", "http://127.0.0.1:8080", "https://pawtnerhope.angelblessy.com"}

	notificationCh = make(chan NotificationJob, 100)
	paymentCh = make(chan Donation, 50)
	paymentConfirmCh = make(chan PaymentConfirmation, 50)
	pendingRegs = make(map[string]*PendingRegistration)

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

	// 2. LOOPING STRUCTURES
	for i, pet := range samplePets {
		pets = append(pets, pet)
		petsByID[pet.ID] = &pets[i]
		statusCounts[pet.Status]++
		petsByBreed[pet.Breed] = append(petsByBreed[pet.Breed], pet.ID)
	}

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
			Duration:    1440,
			Available:   true,
			Features:    []string{"24/7 Care", "Play Area", "Regular Meals"},
		},
	}

	// 2. LOOPING STRUCTURES
	for i := 0; i < len(sampleServices); i++ {
		services = append(services, sampleServices[i])
		servicesByID[sampleServices[i].ID] = &services[i]
		serviceStats[sampleServices[i].ID] = map[string]interface{}{
			"bookings":  0,
			"revenue":   0.0,
			"rating":    4.5,
			"available": sampleServices[i].Available,
		}
	}
}

// 2. CONTROL FLOW
func validatePet(pet Pet) (bool, []string) {
	errs := make([]string, 0)

	if pet.Name == "" {
		errs = append(errs, "Pet name is required")
	}

	if pet.Species == "" {
		errs = append(errs, "Species is required")
	}

	if pet.Age < 0 || pet.Age > 30 {
		errs = append(errs, "Age must be between 0 and 30")
	}

	switch pet.Status {
	case "Available", "Adopted", "Under Care":
	default:
		errs = append(errs, "Invalid status")
	}

	return len(errs) == 0, errs
}

func calculateStatistics() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["petsByStatus"] = statusCounts

	speciesCount := make(map[string]int)
	for _, pet := range pets {
		speciesCount[pet.Species]++
	}
	stats["petsBySpecies"] = speciesCount

	if len(pets) > 0 {
		totalAge := 0
		vaccinatedCount := 0

		// 2. LOOPING
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
	stats["totalDonations"] = len(donations)
	stats["totalInquiries"] = len(inquiries)
	stats["totalUsers"] = len(users)

	return stats
}

// 6. INTERFACE
type SpeciesFilter struct {
	Species string
}

func (f SpeciesFilter) Filter(petList []Pet) []Pet {
	result := make([]Pet, 0)
	for _, p := range petList {
		if strings.EqualFold(p.Species, f.Species) {
			result = append(result, p)
		}
	}
	return result
}

func (f SpeciesFilter) Name() string { return "SpeciesFilter" }

type StatusFilter struct {
	Status string
}

func (f StatusFilter) Filter(petList []Pet) []Pet {
	result := make([]Pet, 0)
	for _, p := range petList {
		if p.Status == f.Status {
			result = append(result, p)
		}
	}
	return result
}

func (f StatusFilter) Name() string { return "StatusFilter" }

type AgeRangeFilter struct {
	Min int
	Max int
}

func (f AgeRangeFilter) Filter(petList []Pet) []Pet {
	result := make([]Pet, 0)
	for _, p := range petList {
		if (f.Min == 0 || p.Age >= f.Min) && (f.Max == 0 || p.Age <= f.Max) {
			result = append(result, p)
		}
	}
	return result
}

func (f AgeRangeFilter) Name() string { return "AgeRangeFilter" }

func ApplyFilters(petList []Pet, filters []Filterable) []Pet {
	result := petList
	for _, filter := range filters {
		result = filter.Filter(result)
	}
	return result
}

// 7. POINTERS, CALL BY VALUE AND CALL BY REFERENCE

func updatePetStatusValue(pet Pet, status string) Pet {
	pet.Status = status
	return pet
}

func updatePetStatusRef(pet *Pet, status string) {
	pet.Status = status
}

func UpdateUserProfile(user *User, username, email string) {
	user.Username = username
	user.Email = email
}

func UpdateToken(token *AuthToken) {
	token.ExpiresAt = time.Now().Add(24 * time.Hour)
}

// 5. FUNCTIONS AND ERROR HANDLING

func hashPassword(password string) string {
	return fmt.Sprintf("hashed_%s_pawtnersalt", password)
}

func generateToken(userID string) string {
	return fmt.Sprintf("tok_%s_%d", userID, time.Now().UnixNano())
}

func Register(email, username, password string) (*User, error) {
	if email == "" || username == "" || password == "" {
		return nil, errors.New("email, username and password are required")
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := usersByEmail[email]; exists {
		return nil, ErrUserAlreadyExists
	}

	user := User{
		ID:        fmt.Sprintf("usr-%03d", len(users)+1),
		Email:     email,
		Username:  username,
		Password:  hashPassword(password),
		Role:      "user",
		CreatedAt: time.Now(),
		IsActive:  true,
	}
	users = append(users, user)
	usersByEmail[email] = &users[len(users)-1]
	return &users[len(users)-1], nil
}

func Login(email, password string) (*AuthToken, error) {
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	mu.Lock()
	defer mu.Unlock()

	user, exists := usersByEmail[email]
	if !exists || user.Password != hashPassword(password) {
		return nil, ErrInvalidCredentials
	}

	token := AuthToken{
		Token:     generateToken(user.ID),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Role:      user.Role,
		IsAdmin:   user.IsAdmin,
		Username:  user.Username,
		Email:     user.Email,
	}
	tokenStore[token.Token] = &token
	return &token, nil
}

func ValidateToken(tokenStr string) (*User, error) {
	if tokenStr == "" {
		return nil, ErrInvalidCredentials
	}

	mu.Lock()
	defer mu.Unlock()

	token, exists := tokenStore[tokenStr]
	if !exists {
		return nil, ErrInvalidCredentials
	}

	if time.Now().After(token.ExpiresAt) {
		delete(tokenStore, tokenStr)
		return nil, ErrTokenExpired
	}

	for i := range users {
		if users[i].ID == token.UserID {
			return &users[i], nil
		}
	}
	return nil, ErrInvalidCredentials
}

func UpdatePet(id string, update Pet) (*Pet, error) {
	mu.Lock()
	defer mu.Unlock()

	pet, exists := petsByID[id]
	if !exists {
		return nil, ErrPetNotFound
	}

	if update.Name != "" {
		pet.Name = update.Name
	}
	if update.Species != "" {
		pet.Species = update.Species
	}
	if update.Breed != "" {
		pet.Breed = update.Breed
	}
	if update.Age > 0 {
		pet.Age = update.Age
	}
	if update.Status != "" {
		oldStatus := pet.Status
		pet.Status = update.Status
		statusCounts[oldStatus]--
		statusCounts[update.Status]++
	}
	if update.Description != "" {
		pet.Description = update.Description
	}
	return pet, nil
}

func DeletePet(id string) error {
	mu.Lock()
	defer mu.Unlock()

	pet, exists := petsByID[id]
	if !exists {
		return ErrPetNotFound
	}

	statusCounts[pet.Status]--
	delete(petsByID, id)

	for i, p := range pets {
		if p.ID == id {
			pets = append(pets[:i], pets[i+1:]...)
			break
		}
	}
	return nil
}

func ProcessDonation(donation *Donation) (*Receipt, error) {
	if donation.Amount <= 0 {
		return nil, ErrInvalidPayment
	}
	if donation.DonorName == "" || donation.DonorEmail == "" {
		return nil, errors.New("donor name and email are required")
	}
	if donation.PaymentMethod == "" {
		return nil, errors.New("payment method is required")
	}

	donation.ID = fmt.Sprintf("don-%03d", len(donations)+1)
	donation.TransactionID = fmt.Sprintf("txn-%d", time.Now().UnixNano())
	donation.Status = "Completed"
	donation.CreatedAt = time.Now()

	mu.Lock()
	donations = append(donations, *donation)
	mu.Unlock()

	syncDonationToDB(*donation)
	receipt := GenerateReceipt(*donation)
	return &receipt, nil
}

func GenerateReceipt(donation Donation) Receipt {
	return Receipt{
		ReceiptID:  fmt.Sprintf("rcpt-%d", time.Now().UnixNano()),
		DonationID: donation.ID,
		DonorName:  donation.DonorName,
		Amount:     donation.Amount,
		IssuedAt:   time.Now(),
		Message:    fmt.Sprintf("Thank you %s for your generous donation of ₹%.2f to Pawtner Hope Foundation!", donation.DonorName, donation.Amount),
	}
}

// ── Email templates ───────────────────────────────────────────────────────────

const welcomeEmailTpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Welcome</title></head>
<body style="margin:0;padding:0;background:#faf8f5;font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#faf8f5;padding:40px 20px;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(44,36,22,.08);">
        <!-- Header -->
        <tr><td style="background:linear-gradient(135deg,#d4a574,#b8844f);padding:40px 48px;text-align:center;">
          <div style="font-size:36px;margin-bottom:8px;">🐾</div>
          <h1 style="margin:0;color:#fff;font-size:26px;font-weight:700;letter-spacing:-0.5px;">Pawtner Hope Foundation</h1>
          <p style="margin:8px 0 0;color:rgba(255,255,255,.8);font-size:14px;">Spreading love, one paw at a time</p>
        </td></tr>
        <!-- Body -->
        <tr><td style="padding:40px 48px;">
          <h2 style="margin:0 0 16px;color:#2c2416;font-size:22px;">Welcome, {{.Username}}! 👋</h2>
          <p style="margin:0 0 16px;color:#555;font-size:15px;line-height:1.7;">Your account has been created successfully. We're so glad to have you as part of our community of animal lovers.</p>
          <p style="margin:0 0 24px;color:#555;font-size:15px;line-height:1.7;">Here's what you can do now:</p>
          <table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
            <tr><td style="padding:12px 16px;background:#fdf6ef;border-left:3px solid #d4a574;border-radius:4px;margin-bottom:10px;">
              <span style="color:#b8844f;font-weight:600;">🐶 Adopt a Pet</span>
              <span style="color:#666;font-size:14px;"> — Browse our animals and submit an adoption inquiry.</span>
            </td></tr>
            <tr><td style="height:8px;"></td></tr>
            <tr><td style="padding:12px 16px;background:#fdf6ef;border-left:3px solid #d4a574;border-radius:4px;">
              <span style="color:#b8844f;font-weight:600;">💛 Donate</span>
              <span style="color:#666;font-size:14px;"> — Help us rescue and care for more animals.</span>
            </td></tr>
          </table>
          <p style="margin:0 0 4px;color:#888;font-size:13px;">Account details</p>
          <table width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #eee;border-radius:8px;overflow:hidden;">
            <tr style="background:#f9f9f9;"><td style="padding:10px 16px;color:#888;font-size:13px;width:120px;">Email</td><td style="padding:10px 16px;color:#2c2416;font-size:13px;">{{.Email}}</td></tr>
            <tr><td style="padding:10px 16px;color:#888;font-size:13px;">Username</td><td style="padding:10px 16px;color:#2c2416;font-size:13px;">{{.Username}}</td></tr>
            <tr style="background:#f9f9f9;"><td style="padding:10px 16px;color:#888;font-size:13px;">Member since</td><td style="padding:10px 16px;color:#2c2416;font-size:13px;">{{.Date}}</td></tr>
          </table>
        </td></tr>
        <!-- CTA -->
        <tr><td style="padding:0 48px 40px;text-align:center;">
          <a href="http://localhost:8080/adoption.html" style="display:inline-block;background:#d4a574;color:#fff;text-decoration:none;padding:14px 36px;border-radius:50px;font-size:15px;font-weight:600;">Browse Pets for Adoption →</a>
        </td></tr>
        <!-- Footer -->
        <tr><td style="background:#f5f0eb;padding:24px 48px;text-align:center;">
          <p style="margin:0;color:#aaa;font-size:12px;">© 2024 Pawtner Hope Foundation. If you didn't create this account, please ignore this email.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`

const receiptEmailTpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Donation Receipt</title></head>
<body style="margin:0;padding:0;background:#faf8f5;font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#faf8f5;padding:40px 20px;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(44,36,22,.08);">
        <!-- Header -->
        <tr><td style="background:linear-gradient(135deg,#d4a574,#b8844f);padding:40px 48px;text-align:center;">
          <div style="font-size:36px;margin-bottom:8px;">🐾</div>
          <h1 style="margin:0;color:#fff;font-size:26px;font-weight:700;">Donation Receipt</h1>
          <p style="margin:8px 0 0;color:rgba(255,255,255,.8);font-size:14px;">Pawtner Hope Foundation</p>
        </td></tr>
        <!-- Amount hero -->
        <tr><td style="padding:36px 48px 24px;text-align:center;border-bottom:1px solid #f0ebe4;">
          <p style="margin:0 0 4px;color:#999;font-size:13px;text-transform:uppercase;letter-spacing:.8px;">Amount Received</p>
          <p style="margin:0;color:#b8844f;font-size:48px;font-weight:700;">₹{{.Amount}}</p>
        </td></tr>
        <!-- Details -->
        <tr><td style="padding:28px 48px;">
          <p style="margin:0 0 16px;color:#2c2416;font-size:16px;font-weight:600;">Thank you, {{.DonorName}}! 💛</p>
          <p style="margin:0 0 24px;color:#555;font-size:15px;line-height:1.7;">Your generous donation helps us rescue, care for, and re-home abandoned pets. Every rupee makes a real difference in an animal's life.</p>
          <table width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #eee;border-radius:10px;overflow:hidden;margin-bottom:24px;">
            <tr style="background:#f9f9f9;"><td style="padding:12px 16px;color:#888;font-size:13px;width:150px;">Receipt No.</td><td style="padding:12px 16px;color:#2c2416;font-size:13px;font-family:monospace;">{{.ReceiptID}}</td></tr>
            <tr><td style="padding:12px 16px;color:#888;font-size:13px;">Donation ID</td><td style="padding:12px 16px;color:#2c2416;font-size:13px;font-family:monospace;">{{.DonationID}}</td></tr>
            <tr style="background:#f9f9f9;"><td style="padding:12px 16px;color:#888;font-size:13px;">UPI Txn / UTR</td><td style="padding:12px 16px;color:#2c2416;font-size:13px;font-family:monospace;">{{.TransactionID}}</td></tr>
            <tr><td style="padding:12px 16px;color:#888;font-size:13px;">Date</td><td style="padding:12px 16px;color:#2c2416;font-size:13px;">{{.Date}}</td></tr>
            <tr style="background:#f9f9f9;"><td style="padding:12px 16px;color:#888;font-size:13px;">Donor Email</td><td style="padding:12px 16px;color:#2c2416;font-size:13px;">{{.DonorEmail}}</td></tr>
          </table>
          <div style="background:#fdf6ef;border-radius:10px;padding:16px 20px;">
            <p style="margin:0;color:#b8844f;font-size:13px;">🔒 This is an official receipt for your tax records. Please save this email.</p>
          </div>
        </td></tr>
        <!-- CTA -->
        <tr><td style="padding:0 48px 40px;text-align:center;">
          <a href="http://localhost:8080/donate.html" style="display:inline-block;background:#d4a574;color:#fff;text-decoration:none;padding:14px 36px;border-radius:50px;font-size:15px;font-weight:600;">Donate Again →</a>
        </td></tr>
        <!-- Footer -->
        <tr><td style="background:#f5f0eb;padding:24px 48px;text-align:center;">
          <p style="margin:0 0 6px;color:#aaa;font-size:12px;">© 2024 Pawtner Hope Foundation</p>
          <p style="margin:0;color:#bbb;font-size:12px;">Questions? Email us at pawtnerhopefoundation@gmail.com</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`

// renderTemplate renders an HTML template string with the given data.
func renderTemplate(tplStr string, data interface{}) (string, error) {
	tpl, err := template.New("").Parse(tplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SendEmail sends an HTML email via Gmail SMTP.
func SendEmail(to, subject, htmlBody string) error {
	if to == "" || subject == "" {
		return ErrEmailFailed
	}
	if emailShouldFail {
		return ErrEmailFailed
	}
	if smtpUser == "" || smtpPass == "" {
		log.Printf("[EMAIL-SKIP] SMTP not configured. To: %s | Subject: %s", to, subject)
		return nil
	}

	header := fmt.Sprintf(
		"From: Pawtner Hope Foundation <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		smtpUser, to, subject,
	)
	message := []byte(header + htmlBody)

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := smtpHost + ":" + smtpPort
	if err := smtp.SendMail(addr, auth, smtpUser, []string{to}, message); err != nil {
		log.Printf("[EMAIL-ERROR] To: %s | %v", to, err)
		return fmt.Errorf("%w: %v", ErrEmailFailed, err)
	}
	log.Printf("[EMAIL-SENT] To: %s | Subject: %s", to, subject)
	return nil
}

func SendEmailWithRetry(to, subject, body string, maxRetries int) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := SendEmail(to, subject, body); err != nil {
			lastErr = err
			log.Printf("[EMAIL] Attempt %d/%d failed for %s: %v", attempt, maxRetries, to, err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}
		return nil
	}
	return fmt.Errorf("email failed after %d attempts: %w", maxRetries, lastErr)
}

// sendWelcomeEmail renders and dispatches the welcome email.
func sendWelcomeEmail(user *User) {
	html, err := renderTemplate(welcomeEmailTpl, map[string]string{
		"Username": user.Username,
		"Email":    user.Email,
		"Date":     user.CreatedAt.Format("2 Jan 2006"),
	})
	if err != nil {
		log.Printf("[EMAIL] Failed to render welcome template: %v", err)
		return
	}
	go SendEmailWithRetry(user.Email, "Welcome to Pawtner Hope Foundation 🐾", html, 3)
}

// sendDonationReceipt renders and dispatches the donation receipt email.
func sendDonationReceipt(donation Donation, receipt Receipt) {
	html, err := renderTemplate(receiptEmailTpl, map[string]string{
		"DonorName":     donation.DonorName,
		"DonorEmail":    donation.DonorEmail,
		"Amount":        fmt.Sprintf("%.2f", donation.Amount),
		"ReceiptID":     receipt.ReceiptID,
		"DonationID":    donation.ID,
		"TransactionID": donation.TransactionID,
		"Date":          donation.CreatedAt.Format("2 Jan 2006, 3:04 PM"),
	})
	if err != nil {
		log.Printf("[EMAIL] Failed to render receipt template: %v", err)
		return
	}
	go SendEmailWithRetry(donation.DonorEmail, "Donation Receipt — Pawtner Hope Foundation 🐾", html, 3)
}

// ── MongoDB helpers ───────────────────────────────────────────────────────────

func petsColl() *mongo.Collection {
	if mongoDB == nil {
		return nil
	}
	return mongoDB.Collection("pets")
}
func usersColl() *mongo.Collection {
	if mongoDB == nil {
		return nil
	}
	return mongoDB.Collection("users")
}
func donationsColl() *mongo.Collection {
	if mongoDB == nil {
		return nil
	}
	return mongoDB.Collection("donations")
}
func inquiriesColl() *mongo.Collection {
	if mongoDB == nil {
		return nil
	}
	return mongoDB.Collection("inquiries")
}

func syncPetToDB(pet Pet) {
	if petsColl() == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := options.Replace().SetUpsert(true)
		if _, err := petsColl().ReplaceOne(ctx, bson.M{"id": pet.ID}, pet, opts); err != nil {
			log.Printf("[MONGO] syncPetToDB error: %v", err)
		}
	}()
}

func deletePetFromDB(petID string) {
	if petsColl() == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := petsColl().DeleteOne(ctx, bson.M{"id": petID}); err != nil {
			log.Printf("[MONGO] deletePetFromDB error: %v", err)
		}
	}()
}

func syncUserToDB(user User) {
	if usersColl() == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := options.Replace().SetUpsert(true)
		if _, err := usersColl().ReplaceOne(ctx, bson.M{"id": user.ID}, user, opts); err != nil {
			log.Printf("[MONGO] syncUserToDB error: %v", err)
		}
	}()
}

func syncDonationToDB(donation Donation) {
	if donationsColl() == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := options.Replace().SetUpsert(true)
		if _, err := donationsColl().ReplaceOne(ctx, bson.M{"id": donation.ID}, donation, opts); err != nil {
			log.Printf("[MONGO] syncDonationToDB error: %v", err)
		}
	}()
}

func syncInquiryToDB(inquiry AdoptionInquiry) {
	if inquiriesColl() == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := options.Replace().SetUpsert(true)
		if _, err := inquiriesColl().ReplaceOne(ctx, bson.M{"id": inquiry.ID}, inquiry, opts); err != nil {
			log.Printf("[MONGO] syncInquiryToDB error: %v", err)
		}
	}()
}

// loadFromMongoDB seeds in-memory data from MongoDB collections on startup.
// If a collection is empty it falls back to whatever initializeData() put there.
func loadFromMongoDB() {
	if mongoDB == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Pets
	if cur, err := petsColl().Find(ctx, bson.D{}); err == nil {
		var dbPets []Pet
		if err := cur.All(ctx, &dbPets); err == nil && len(dbPets) > 0 {
			mu.Lock()
			pets = dbPets
			petsByID = make(map[string]*Pet)
			statusCounts = make(map[string]int)
			petsByBreed = make(map[string][]string)
			for i := range pets {
				petsByID[pets[i].ID] = &pets[i]
				statusCounts[pets[i].Status]++
				petsByBreed[pets[i].Breed] = append(petsByBreed[pets[i].Breed], pets[i].ID)
			}
			mu.Unlock()
			log.Printf("[MONGO] Loaded %d pets", len(pets))
		} else if err == nil {
			// Collection is empty — push sample data to MongoDB so it persists
			log.Println("[MONGO] No pets in DB, seeding sample data")
			for _, p := range pets {
				syncPetToDB(p)
			}
		}
	}

	// Users
	if cur, err := usersColl().Find(ctx, bson.D{}); err == nil {
		var dbUsers []User
		if err := cur.All(ctx, &dbUsers); err == nil && len(dbUsers) > 0 {
			mu.Lock()
			users = dbUsers
			usersByEmail = make(map[string]*User)
			for i := range users {
				usersByEmail[users[i].Email] = &users[i]
			}
			mu.Unlock()
			log.Printf("[MONGO] Loaded %d users", len(users))
		}
	}

	// Donations
	if cur, err := donationsColl().Find(ctx, bson.D{}); err == nil {
		var dbDonations []Donation
		if err := cur.All(ctx, &dbDonations); err == nil && len(dbDonations) > 0 {
			mu.Lock()
			donations = dbDonations
			mu.Unlock()
			log.Printf("[MONGO] Loaded %d donations", len(donations))
		}
	}

	// Inquiries
	if cur, err := inquiriesColl().Find(ctx, bson.D{}); err == nil {
		var dbInquiries []AdoptionInquiry
		if err := cur.All(ctx, &dbInquiries); err == nil && len(dbInquiries) > 0 {
			mu.Lock()
			inquiries = dbInquiries
			mu.Unlock()
			log.Printf("[MONGO] Loaded %d inquiries", len(inquiries))
		}
	}
}

// generateOTP returns a zero-padded 6-digit numeric code.
func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// ── OTP email template ────────────────────────────────────────────────────────

const otpEmailTpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Verify Email</title></head>
<body style="margin:0;padding:0;background:#faf8f5;font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#faf8f5;padding:40px 20px;">
    <tr><td align="center">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(44,36,22,.08);">
        <tr><td style="background:linear-gradient(135deg,#d4a574,#b8844f);padding:36px 48px;text-align:center;">
          <div style="font-size:36px;margin-bottom:8px;">🐾</div>
          <h1 style="margin:0;color:#fff;font-size:24px;font-weight:700;">Email Verification</h1>
          <p style="margin:8px 0 0;color:rgba(255,255,255,.8);font-size:14px;">Pawtner Hope Foundation</p>
        </td></tr>
        <tr><td style="padding:40px 48px;text-align:center;">
          <p style="margin:0 0 8px;color:#555;font-size:15px;line-height:1.7;">Hi <strong>{{.Username}}</strong>! Use the code below to verify your email address.</p>
          <p style="margin:0 0 28px;color:#888;font-size:13px;">This code expires in <strong>5 minutes</strong>.</p>
          <div style="display:inline-block;background:#fdf6ef;border:2px dashed #d4a574;border-radius:16px;padding:24px 48px;margin-bottom:28px;">
            <p style="margin:0;font-size:42px;font-weight:800;letter-spacing:10px;color:#b8844f;font-family:monospace;">{{.Code}}</p>
          </div>
          <p style="margin:0;color:#aaa;font-size:12px;">If you didn't request this, you can safely ignore this email.</p>
        </td></tr>
        <tr><td style="background:#f5f0eb;padding:20px 48px;text-align:center;">
          <p style="margin:0;color:#aaa;font-size:12px;">© 2024 Pawtner Hope Foundation</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body></html>`

// 5. FUNCTIONS AND ERROR HANDLING
func SearchPets(query string, filters []Filterable) ([]Pet, error) {
	if query == "" && len(filters) == 0 {
		return nil, errors.New("search query or filters required")
	}

	mu.Lock()
	petsCopy := make([]Pet, len(pets))
	copy(petsCopy, pets)
	mu.Unlock()

	var result []Pet
	if query != "" {
		for _, p := range petsCopy {
			nameLower := strings.ToLower(p.Name)
			speciesLower := strings.ToLower(p.Species)
			breedLower := strings.ToLower(p.Breed)
			qLower := strings.ToLower(query)
			if strings.Contains(nameLower, qLower) ||
				strings.Contains(speciesLower, qLower) ||
				strings.Contains(breedLower, qLower) {
				result = append(result, p)
			}
		}
	} else {
		result = petsCopy
	}

	if len(filters) > 0 {
		result = ApplyFilters(result, filters)
	}

	return result, nil
}

// 8. JSON MARSHAL AND UNMARSHAL

func MarshalPet(pet Pet) ([]byte, error) {
	return json.Marshal(pet)
}

func UnmarshalPet(data []byte) (*Pet, error) {
	var pet Pet
	if err := json.Unmarshal(data, &pet); err != nil {
		return nil, err
	}
	return &pet, nil
}

func MarshalDonation(donation Donation) ([]byte, error) {
	return json.Marshal(donation)
}

func UnmarshalDonation(data []byte) (*Donation, error) {
	var donation Donation
	if err := json.Unmarshal(data, &donation); err != nil {
		return nil, err
	}
	return &donation, nil
}

// 10. CONCURRENCY

// 11. GOROUTINES AND CHANNELS
func emailWorker(jobs <-chan NotificationJob) {
	for job := range jobs {
		SendEmailWithRetry(job.To, job.Subject, job.Body, 3)
	}
}

func paymentProcessor(donationQueue <-chan Donation, confirmations chan<- PaymentConfirmation) {
	for donation := range donationQueue {
		time.Sleep(50 * time.Millisecond)
		confirmation := PaymentConfirmation{
			DonationID:    donation.ID,
			Success:       true,
			TransactionID: fmt.Sprintf("txn-%d", time.Now().UnixNano()),
		}
		confirmations <- confirmation

		// Only auto-send receipt for mobile UPI deeplink payments.
		// Desktop donors must request a receipt via email.
		if donation.PaymentViaDeeplink {
			go func(d Donation) {
				receipt := GenerateReceipt(d)
				sendDonationReceipt(d, receipt)
			}(donation)
		} else {
			log.Printf("[INFO] Desktop donation from %s — receipt not auto-sent (request required)", donation.DonorEmail)
		}
	}
}

func confirmationListener(confirmations <-chan PaymentConfirmation) {
	for confirmation := range confirmations {
		mu.Lock()
		for i := range donations {
			if donations[i].ID == confirmation.DonationID {
				if confirmation.Success {
					donations[i].Status = "Completed"
					donations[i].TransactionID = confirmation.TransactionID
				} else {
					donations[i].Status = "Failed"
				}
				break
			}
		}
		mu.Unlock()
		log.Printf("[PAYMENT] Processed: %s - Success: %v", confirmation.DonationID, confirmation.Success)
	}
}

func parallelSearch(criteria map[string]string) []Pet {
	type searchResult struct {
		pets []Pet
	}
	resultCh := make(chan searchResult, len(criteria))

	mu.Lock()
	petsCopy := make([]Pet, len(pets))
	copy(petsCopy, pets)
	mu.Unlock()

	var wg sync.WaitGroup
	for key, value := range criteria {
		wg.Add(1)
		go func(k, v string) {
			defer wg.Done()
			var filters []Filterable
			switch k {
			case "species":
				filters = append(filters, SpeciesFilter{Species: v})
			case "status":
				filters = append(filters, StatusFilter{Status: v})
			}
			resultCh <- searchResult{pets: ApplyFilters(petsCopy, filters)}
		}(key, value)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	seen := make(map[string]bool)
	var combined []Pet
	for r := range resultCh {
		for _, p := range r.pets {
			if !seen[p.ID] {
				seen[p.ID] = true
				combined = append(combined, p)
			}
		}
	}
	return combined
}

func startWorkers() {
	// 11. GOROUTINES AND CHANNELS
	go emailWorker(notificationCh)
	go paymentProcessor(paymentCh, paymentConfirmCh)
	go confirmationListener(paymentConfirmCh)
}

// HTTP Handlers

// Panic recovery middleware
func recoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v for request %s %s", err, r.Method, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "Internal server error",
				})
			}
		}()
		next(w, r)
	}
}

// Request logging middleware
func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[REQUEST] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
		log.Printf("[RESPONSE] %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

// 6. INTERFACE - http.HandlerFunc implements http.Handler
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

// Safe file serving with error handling
func serveHTMLFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Printf("[ERROR] File not found: %s", filename)
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, filename)
	}
}

// Safe JSON response with error handling
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] Failed to encode JSON response: %v", err)
	}
}

// Error response helper
func respondError(w http.ResponseWriter, statusCode int, message string) {
	log.Printf("[ERROR] HTTP %d: %s", statusCode, message)
	respondJSON(w, statusCode, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// Input sanitization helper (basic XSS prevention)
func sanitizeString(s string) string {
	// Remove potentially harmful characters
	s = strings.TrimSpace(s)
	// Basic HTML entity encoding for display in web context
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func getPetsHandler(w http.ResponseWriter, r *http.Request) {
	requestCount++

	query := r.URL.Query()
	species := query.Get("species")
	status := query.Get("status")
	search := query.Get("q")

	var result []Pet

	// 2. CONTROL FLOW
	if search != "" {
		var filters []Filterable
		if species != "" {
			filters = append(filters, SpeciesFilter{Species: species})
		}
		if status != "" {
			filters = append(filters, StatusFilter{Status: status})
		}
		var err error
		result, err = SearchPets(search, filters)
		if err != nil {
			result = pets
		}
	} else if species == "" && status == "" {
		result = pets
	} else {
		var filters []Filterable
		if species != "" {
			filters = append(filters, SpeciesFilter{Species: species})
		}
		if status != "" {
			filters = append(filters, StatusFilter{Status: status})
		}
		result = ApplyFilters(pets, filters)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(result),
		"data":    result,
	})
}

func getPetByIDHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/pets/")
	petID := path

	pet, exists := petsByID[petID]

	// 2. CONTROL FLOW
	if !exists {
		respondError(w, http.StatusNotFound, "Pet not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    pet,
	})
}

func addPetHandler(w http.ResponseWriter, r *http.Request) {
	var newPet Pet

	// 8. JSON MARSHAL AND UNMARSHAL
	if err := json.NewDecoder(r.Body).Decode(&newPet); err != nil {
		log.Printf("[ERROR] Failed to decode pet JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON data")
		return
	}
	defer r.Body.Close()

	valid, validationErrors := validatePet(newPet)
	if !valid {
		log.Printf("[ERROR] Pet validation failed: %v", validationErrors)
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Validation failed",
			"errors":  validationErrors,
		})
		return
	}

	newPet.ID = fmt.Sprintf("pet-%03d", len(pets)+1)
	newPet.CreatedAt = time.Now()

	mu.Lock()
	pets = append(pets, newPet)
	petsByID[newPet.ID] = &pets[len(pets)-1]
	statusCounts[newPet.Status]++
	petsByBreed[newPet.Breed] = append(petsByBreed[newPet.Breed], newPet.ID)
	mu.Unlock()

	syncPetToDB(newPet)
	log.Printf("[INFO] Pet added: ID=%s, Name=%s, Species=%s", newPet.ID, newPet.Name, newPet.Species)
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Pet added successfully",
		"data":    newPet,
	})
}

func updatePetHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/pets/")
	petID := path

	var update Pet

	// 8. JSON MARSHAL AND UNMARSHAL
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("[ERROR] Failed to decode update JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON data")
		return
	}
	defer r.Body.Close()

	// 5. FUNCTIONS AND ERROR HANDLING
	pet, err := UpdatePet(petID, update)
	if err != nil {
		if errors.Is(err, ErrPetNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	log.Printf("[INFO] Pet updated: ID=%s", petID)
	syncPetToDB(*pet)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Pet updated successfully",
		"data":    pet,
	})
}

func deletePetHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/pets/")
	petID := path

	// 5. FUNCTIONS AND ERROR HANDLING
	if err := DeletePet(petID); err != nil {
		if errors.Is(err, ErrPetNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.Printf("[INFO] Pet deleted: ID=%s", petID)
	deletePetFromDB(petID)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Pet deleted successfully",
	})
}

func getServicesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	category := query.Get("category")

	var result []Service

	// 2. CONTROL FLOW and LOOPING
	if category == "" {
		result = services
	} else {
		for _, service := range services {
			if service.Category == category {
				result = append(result, service)
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(result),
		"data":    result,
	})
}

func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	var booking ServiceBooking

	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		log.Printf("[ERROR] Failed to decode booking JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON data")
		return
	}
	defer r.Body.Close()

	// Validate required fields
	if booking.ServiceID == "" || booking.OwnerName == "" || booking.Email == "" {
		respondError(w, http.StatusBadRequest, "Service ID, owner name, and email are required")
		return
	}

	booking.ID = fmt.Sprintf("book-%03d", len(bookings)+1)
	booking.BookedAt = time.Now()
	booking.Status = "Pending"

	mu.Lock()
	bookings = append(bookings, booking)
	bookingsByID[booking.ID] = &bookings[len(bookings)-1]
	if stats, exists := serviceStats[booking.ServiceID]; exists {
		stats["bookings"] = stats["bookings"].(int) + 1
	}
	mu.Unlock()

	log.Printf("[INFO] Booking created: ID=%s, Service=%s, Owner=%s", booking.ID, booking.ServiceID, booking.OwnerName)
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Booking created successfully",
		"data":    booking,
	})
}

func submitContactHandler(w http.ResponseWriter, r *http.Request) {
	var contact ContactForm

	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		log.Printf("[ERROR] Failed to decode contact JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON data")
		return
	}
	defer r.Body.Close()

	// Validate required fields
	if contact.Name == "" || contact.Email == "" || contact.Message == "" {
		respondError(w, http.StatusBadRequest, "Name, email, and message are required")
		return
	}

	contact.SentAt = time.Now()
	mu.Lock()
	contactMessages = append(contactMessages, contact)
	mu.Unlock()

	log.Printf("[INFO] Contact message received from: %s (%s)", contact.Name, contact.Email)

	// 10. CONCURRENCY
	go func() {
		notificationCh <- NotificationJob{
			To:      contact.Email,
			Subject: "Thank you for contacting Pawtner Hope",
			Body:    fmt.Sprintf("Dear %s, we received your message and will get back to you soon.", contact.Name),
			JobType: "contact",
		}
	}()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Message sent successfully",
	})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode registration JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	if req.Email == "" || req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email, username and password are required")
		return
	}

	mu.Lock()
	_, alreadyExists := usersByEmail[req.Email]
	mu.Unlock()
	if alreadyExists {
		respondError(w, http.StatusConflict, ErrUserAlreadyExists.Error())
		return
	}

	code := generateOTP()
	pending := &PendingRegistration{
		Email:          req.Email,
		Username:       req.Username,
		HashedPassword: hashPassword(req.Password),
		Code:           code,
		ExpiresAt:      time.Now().Add(5 * time.Minute),
	}
	mu.Lock()
	pendingRegs[req.Email] = pending
	mu.Unlock()

	// Send OTP email asynchronously
	go func() {
		html, err := renderTemplate(otpEmailTpl, map[string]string{
			"Username": req.Username,
			"Code":     code,
		})
		if err != nil {
			log.Printf("[EMAIL] Failed to render OTP template: %v", err)
			return
		}
		SendEmailWithRetry(req.Email, "Your Pawtner Hope Verification Code 🐾", html, 3)
	}()

	log.Printf("[INFO] OTP sent to %s (expires in 5 min)", req.Email)
	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"success": true,
		"message": "Verification code sent to your email. It expires in 5 minutes.",
	})
}

func verifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Code = strings.TrimSpace(req.Code)

	mu.Lock()
	pending, exists := pendingRegs[req.Email]
	mu.Unlock()

	if !exists {
		respondError(w, http.StatusBadRequest, "No pending registration for this email. Please sign up again.")
		return
	}
	if time.Now().After(pending.ExpiresAt) {
		mu.Lock()
		delete(pendingRegs, req.Email)
		mu.Unlock()
		respondError(w, http.StatusBadRequest, "Verification code has expired. Please sign up again.")
		return
	}
	if req.Code != pending.Code {
		respondError(w, http.StatusBadRequest, "Invalid verification code.")
		return
	}

	// Create user with pre-hashed password
	user := User{
		ID:        fmt.Sprintf("usr-%03d", len(users)+1),
		Email:     pending.Email,
		Username:  pending.Username,
		Password:  pending.HashedPassword,
		Role:      "user",
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	mu.Lock()
	users = append(users, user)
	usersByEmail[user.Email] = &users[len(users)-1]
	delete(pendingRegs, req.Email)
	mu.Unlock()

	syncUserToDB(user)
	sendWelcomeEmail(&user)
	log.Printf("[INFO] User verified and created: %s (%s)", user.Username, user.Email)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Account created successfully! You can now log in.",
		"data":    user,
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode login JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	// 5. FUNCTIONS AND ERROR HANDLING
	token, err := Login(req.Email, req.Password)
	if err != nil {
		log.Printf("[WARN] Failed login attempt for: %s", req.Email)
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	log.Printf("[INFO] User logged in: %s", req.Email)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"data":    token,
	})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		respondError(w, http.StatusUnauthorized, "Missing token")
		return
	}
	user, err := ValidateToken(tokenStr)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"username":  user.Username,
			"role":      user.Role,
			"isAdmin":   user.IsAdmin,
			"createdAt": user.CreatedAt,
		},
	})
}

func createAdoptionInquiryHandler(w http.ResponseWriter, r *http.Request) {
	var inquiry AdoptionInquiry

	// 8. JSON MARSHAL AND UNMARSHAL
	if err := json.NewDecoder(r.Body).Decode(&inquiry); err != nil {
		log.Printf("[ERROR] Failed to decode adoption inquiry JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	if inquiry.PetID == "" || inquiry.AdopterName == "" || inquiry.Email == "" {
		respondError(w, http.StatusBadRequest, "PetID, adopter name and email are required")
		return
	}

	inquiry.ID = fmt.Sprintf("inq-%03d", len(inquiries)+1)
	inquiry.Status = "Pending"
	inquiry.CreatedAt = time.Now()

	mu.Lock()
	inquiries = append(inquiries, inquiry)
	mu.Unlock()

	syncInquiryToDB(inquiry)
	log.Printf("[INFO] Adoption inquiry: Pet=%s, Adopter=%s (%s)", inquiry.PetID, inquiry.AdopterName, inquiry.Email)

	// 10. CONCURRENCY
	go func() {
		notificationCh <- NotificationJob{
			To:      inquiry.Email,
			Subject: "Adoption Inquiry Received - Pawtner Hope",
			Body:    fmt.Sprintf("Dear %s, your adoption inquiry for pet %s has been received.", inquiry.AdopterName, inquiry.PetID),
			JobType: "adoption",
		}
	}()

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Adoption inquiry submitted successfully",
		"data":    inquiry,
	})
}

func getAdoptionInquiriesHandler(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	result := make([]AdoptionInquiry, len(inquiries))
	copy(result, inquiries)
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(result),
		"data":    result,
	})
}

func createDonationHandler(w http.ResponseWriter, r *http.Request) {
	var donation Donation

	// 8. JSON MARSHAL AND UNMARSHAL
	if err := json.NewDecoder(r.Body).Decode(&donation); err != nil {
		log.Printf("[ERROR] Failed to decode donation JSON: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	// 5. FUNCTIONS AND ERROR HANDLING
	receipt, err := ProcessDonation(&donation)
	if err != nil {
		log.Printf("[ERROR] Donation processing failed: %v", err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[INFO] Donation received: ₹%.2f from %s (%s) deeplink=%v",
		donation.Amount, donation.DonorName, donation.DonorEmail, donation.PaymentViaDeeplink)

	// 11. GOROUTINES AND CHANNELS — send to payment processor
	go func(d Donation) {
		paymentCh <- d
	}(donation)

	receiptHint := ""
	if !donation.PaymentViaDeeplink {
		receiptHint = "Donation recorded. A receipt can be requested by emailing pawtnerhopefoundation@gmail.com."
	} else {
		receiptHint = "Donation recorded. A receipt has been sent to " + donation.DonorEmail + "."
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": receiptHint,
		"data":    receipt,
	})
}

func getDonationsHandler(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	result := make([]Donation, len(donations))
	copy(result, donations)
	mu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(result),
		"data":    result,
	})
}

func getStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	stats := calculateStatistics()
	stats["serverVersion"] = serverVersion
	stats["uptime"] = time.Since(serverStartTime).String()
	stats["requestCount"] = requestCount
	stats["isProduction"] = isProduction
	stats["serviceStats"] = serviceStats

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

func main() {
	// Load .env before anything else so SMTP credentials are available.
	loadEnv(".env")
	smtpUser = os.Getenv("GMAIL_USER")
	smtpPass = os.Getenv("GMAIL_PASS")
	if smtpUser != "" {
		log.Printf("[SMTP] Email configured for: %s", smtpUser)
	} else {
		log.Println("[SMTP] No GMAIL_USER set \u2014 emails will be skipped")
	}

	initializeData()
	startWorkers()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("⚠ MONGODB_URI not set, running without database")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

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

			if err := client.Ping(ctx, nil); err != nil {
				log.Printf("Failed to ping MongoDB: %v", err)
			} else {
				log.Println("✓ Successfully connected to MongoDB!")
				mongoClient = client
				mongoDB = client.Database("pawtner-hope")
				loadFromMongoDB()
			}
		}
	}

	// Serve HTML files with error handling
	http.HandleFunc("/", recoverPanic(serveHTMLFile("index.html")))
	http.HandleFunc("/about", recoverPanic(serveHTMLFile("index.html")))
	http.HandleFunc("/service.html", recoverPanic(serveHTMLFile("service.html")))
	http.HandleFunc("/adoption.html", recoverPanic(serveHTMLFile("adoption.html")))
	http.HandleFunc("/donate.html", recoverPanic(serveHTMLFile("donate.html")))
	http.HandleFunc("/auth.html", recoverPanic(serveHTMLFile("auth.html")))
	http.HandleFunc("/admin.html", recoverPanic(serveHTMLFile("admin.html")))
	http.HandleFunc("/dashboard.html", recoverPanic(serveHTMLFile("dashboard.html")))

	http.HandleFunc("/api/pets", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		// 2. CONTROL FLOW
		switch r.Method {
		case "GET":
			getPetsHandler(w, r)
		case "POST":
			addPetHandler(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	http.HandleFunc("/api/pets/", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getPetByIDHandler(w, r)
		case "PUT":
			updatePetHandler(w, r)
		case "DELETE":
			deletePetHandler(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	http.HandleFunc("/api/services", recoverPanic(enableCORS(getServicesHandler)))
	http.HandleFunc("/api/bookings", recoverPanic(enableCORS(createBookingHandler)))
	http.HandleFunc("/api/contact", recoverPanic(enableCORS(submitContactHandler)))
	http.HandleFunc("/api/statistics", recoverPanic(enableCORS(getStatisticsHandler)))

	http.HandleFunc("/api/auth/register", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			registerHandler(w, r)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	http.HandleFunc("/api/auth/login", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			loginHandler(w, r)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	http.HandleFunc("/api/auth/verify", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			verifyEmailHandler(w, r)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	http.HandleFunc("/api/auth/me", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			meHandler(w, r)
		} else {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	http.HandleFunc("/api/adoptions", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getAdoptionInquiriesHandler(w, r)
		case "POST":
			createAdoptionInquiryHandler(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	http.HandleFunc("/api/donations", recoverPanic(enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getDonationsHandler(w, r)
		case "POST":
			createDonationHandler(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})))

	log.Println("==============================================")
	log.Println("🐾 Pawtner Hope Foundation Server")
	log.Println("==============================================")
	log.Printf("Version: %s\n", serverVersion)
	log.Printf("Initialized with %d pets\n", len(pets))
	log.Printf("Initialized with %d services\n", len(services))
	log.Println("==============================================")
	log.Println("API Endpoints:")
	log.Println("  GET    /api/pets              - Get all pets")
	log.Println("  GET    /api/pets/:id          - Get pet by ID")
	log.Println("  POST   /api/pets              - Add new pet")
	log.Println("  PUT    /api/pets/:id          - Update pet")
	log.Println("  DELETE /api/pets/:id          - Delete pet")
	log.Println("  GET    /api/services          - Get all services")
	log.Println("  POST   /api/bookings          - Create booking")
	log.Println("  POST   /api/contact           - Submit contact form")
	log.Println("  GET    /api/statistics        - Get statistics")
	log.Println("  POST   /api/auth/register     - Register user")
	log.Println("  POST   /api/auth/login        - Login user")
	log.Println("  GET    /api/adoptions         - Get adoption inquiries")
	log.Println("  POST   /api/adoptions         - Submit adoption inquiry")
	log.Println("  GET    /api/donations         - Get donations")
	log.Println("  POST   /api/donations         - Process donation")
	log.Println("==============================================")
	log.Println("Server starting on http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

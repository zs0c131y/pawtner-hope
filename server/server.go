package main

import (
	"fmt"
)

// 1. VARIABLES, VALUES AND TYPES

// Variables with explicit type declaration
var (
	// Basic data types
	name        string  = "Buddy" // String type - text data
	age         int     = 3       // Integer type - whole numbers
	price       float64 = 99.99   // Float type - decimal numbers
	isAvailable bool    = true    // Boolean type - true/false values
	weight      float32 = 25.5    // 32-bit float
	count       int32   = 100     // 32-bit integer
	id          uint    = 12345   // Unsigned integer (positive only)
	grade       byte    = 'A'     // Byte type (alias for uint8)
	score       rune    = '★'     // Rune type (alias for int32, Unicode code point)
)

// 2. STRUCTS

// Pet struct - demonstrates custom data type with multiple fields
type Pet struct {
	ID           string            // Unique identifier
	Name         string            // Pet name
	Species      string            // Type of animal (Dog, Cat, etc.)
	Age          int               // Age in years
	IsVaccinated bool              // Vaccination status
	Tags         []string          // Slice of tags (dynamic array)
	Attributes   map[string]string // Map of key-value pairs
}

// Owner struct - another example of struct
type Owner struct {
	Name    string // Owner name
	Email   string // Contact email
	Phone   string // Contact phone
	PetList []Pet  // List of pets owned (slice of Pet structs)
}

//  3. ARRAYS AND SLICES

// Arrays - fixed size collections
var (
	// Array: Fixed-size collection (size cannot change)
	petTypes [4]string  = [4]string{"Dog", "Cat", "Bird", "Rabbit"} // Array of 4 strings
	ratings  [5]int     = [5]int{5, 4, 5, 3, 4}                     // Array of 5 integers
	prices   [3]float64                                             // Array initialized to zero values
)

// Slices - dynamic size collections
var (
	// Slice: Dynamic-size collection (can grow or shrink)
	availablePets []string                                       // Slice declaration (nil initially)
	adoptedPets   []string = []string{}                          // Empty slice
	petNames      []string = []string{"Max", "Bella", "Charlie"} // Slice with initial values
	petAges       []int                                          // Slice of integers
	services      []string                                       // Slice for services
)

// 4. MAPS

// Maps - key-value pairs (like dictionary or hashtable)
var (
	// Map: Collection of key-value pairs
	petByID        map[string]string         // Map with string keys and string values
	ageByName      map[string]int            // Map with string keys and int values
	priceByService map[string]float64        // Map for service prices
	petDetails     map[string]Pet            // Map with Pet struct as value
	nestedMap      map[string]map[string]int // Nested map
	statusCount    map[string]int            // Count by status
)

//  MAIN FUNCTION

func main() {
	fmt.Println("=== Go Programming Concepts Demo ===")

	//  VARIABLES DEMONSTRATION
	fmt.Println("--- 1. Variables, Values and Types ---")
	fmt.Printf("String: %s (type: %T)\n", name, name)
	fmt.Printf("Integer: %d (type: %T)\n", age, age)
	fmt.Printf("Float64: %.2f (type: %T)\n", price, price)
	fmt.Printf("Boolean: %v (type: %T)\n", isAvailable, isAvailable)
	fmt.Printf("Float32: %.1f (type: %T)\n", weight, weight)
	fmt.Printf("Byte: %c (type: %T)\n", grade, grade)
	fmt.Printf("Rune: %c (type: %T)\n\n", score, score)

	//  ARRAYS DEMONSTRATION
	fmt.Println("--- 2. Arrays (Fixed Size) ---")
	fmt.Printf("Pet Types Array: %v\n", petTypes)
	fmt.Printf("Array Length: %d\n", len(petTypes))
	fmt.Printf("First Element: %s\n", petTypes[0])
	fmt.Printf("Last Element: %s\n\n", petTypes[len(petTypes)-1])

	// SLICES DEMONSTRATION
	fmt.Println("--- 3. Slices (Dynamic Size) ---")

	// Appending to slices
	availablePets = append(availablePets, "Max", "Buddy", "Luna")
	fmt.Printf("Available Pets: %v\n", availablePets)
	fmt.Printf("Slice Length: %d, Capacity: %d\n", len(availablePets), cap(availablePets))

	// Appending more elements
	availablePets = append(availablePets, "Charlie", "Bella")
	fmt.Printf("After Append: %v\n", availablePets)
	fmt.Printf("Slice Length: %d, Capacity: %d\n", len(availablePets), cap(availablePets))

	// Slice operations
	fmt.Printf("First Three: %v\n", availablePets[:3])
	fmt.Printf("From Index 2: %v\n", availablePets[2:])
	fmt.Printf("Range [1:4]: %v\n\n", availablePets[1:4])

	//  MAPS DEMONSTRATION
	fmt.Println("--- 4. Maps (Key-Value Pairs) ---")

	// Initialize maps using make
	petByID = make(map[string]string)
	ageByName = make(map[string]int)
	priceByService = make(map[string]float64)

	// Adding key-value pairs
	petByID["P001"] = "Max"
	petByID["P002"] = "Bella"
	petByID["P003"] = "Charlie"

	ageByName["Max"] = 3
	ageByName["Bella"] = 5
	ageByName["Charlie"] = 2

	priceByService["Grooming"] = 45.99
	priceByService["Vaccination"] = 75.00
	priceByService["Checkup"] = 50.00

	fmt.Printf("Pet By ID: %v\n", petByID)
	fmt.Printf("Pet P001: %s\n", petByID["P001"])
	fmt.Printf("Age By Name: %v\n", ageByName)
	fmt.Printf("Max's Age: %d\n", ageByName["Max"])
	fmt.Printf("Service Prices: %v\n\n", priceByService)

	// STRUCTS DEMONSTRATION
	fmt.Println("--- 5. Structs (Custom Types) ---")

	// Creating a Pet struct instance
	pet1 := Pet{
		ID:           "P001",
		Name:         "Max",
		Species:      "Dog",
		Age:          3,
		IsVaccinated: true,
		Tags:         []string{"friendly", "trained", "energetic"},
		Attributes: map[string]string{
			"color": "Golden",
			"size":  "Medium",
		},
	}

	fmt.Printf("Pet: %+v\n", pet1)
	fmt.Printf("Pet Name: %s\n", pet1.Name)
	fmt.Printf("Pet Tags: %v\n", pet1.Tags)
	fmt.Printf("Pet Attributes: %v\n\n", pet1.Attributes)

	//  LOOPING STRUCTURES
	fmt.Println("--- 6. Looping Structures ---")

	// For loop - basic counter
	fmt.Println("Basic For Loop (Counter):")
	for i := 0; i < 5; i++ {
		fmt.Printf("  Count: %d\n", i)
	}

	// For loop - range over slice
	fmt.Println("\nFor-Range Loop (Slice):")
	for index, pet := range availablePets {
		fmt.Printf("  [%d] %s\n", index, pet)
	}

	// For loop - range over array
	fmt.Println("\nFor-Range Loop (Array):")
	for index, petType := range petTypes {
		fmt.Printf("  [%d] %s\n", index, petType)
	}

	// For loop - range over map
	fmt.Println("\nFor-Range Loop (Map):")
	for key, value := range ageByName {
		fmt.Printf("  %s is %d years old\n", key, value)
	}

	// While-style loop (for with condition only)
	fmt.Println("\nWhile-Style Loop:")
	counter := 0
	for counter < 3 {
		fmt.Printf("  Counter: %d\n", counter)
		counter++
	}

	// Infinite loop with break
	fmt.Println("\nInfinite Loop with Break:")
	num := 0
	for {
		if num >= 3 {
			break // Exit loop when condition is met
		}
		fmt.Printf("  Number: %d\n", num)
		num++
	}

	// Loop with continue
	fmt.Println("\nLoop with Continue:")
	for i := 0; i < 5; i++ {
		if i == 2 {
			continue // Skip iteration when i is 2
		}
		fmt.Printf("  Value: %d\n", i)
	}

	//  CONTROL FLOW
	fmt.Println("\n--- 7. Control Flow ---")

	// If-else statement
	fmt.Println("If-Else Statements:")
	petAge := 3
	if petAge < 2 {
		fmt.Println("  Pet is a puppy/kitten")
	} else if petAge < 7 {
		fmt.Println("  Pet is an adult")
	} else {
		fmt.Println("  Pet is a senior")
	}

	// If with short statement
	fmt.Println("\nIf with Short Statement:")
	if currentAge := 5; currentAge >= 7 {
		fmt.Printf("  Senior pet (age: %d)\n", currentAge)
	} else {
		fmt.Printf("  Not a senior yet (age: %d)\n", currentAge)
	}

	// Switch statement
	fmt.Println("\nSwitch Statement:")
	species := "Dog"
	switch species {
	case "Dog":
		fmt.Println("  Woof! This is a dog")
	case "Cat":
		fmt.Println("  Meow! This is a cat")
	case "Bird":
		fmt.Println("  Tweet! This is a bird")
	default:
		fmt.Println("  Unknown species")
	}

	// Switch with multiple cases
	fmt.Println("\nSwitch with Multiple Cases:")
	day := 3
	switch day {
	case 1, 2, 3, 4, 5:
		fmt.Println("  Weekday")
	case 6, 7:
		fmt.Println("  Weekend")
	default:
		fmt.Println("  Invalid day")
	}

	// Switch without expression (like if-else chain)
	fmt.Println("\nSwitch without Expression:")
	temperature := 25
	switch {
	case temperature < 0:
		fmt.Println("  Freezing")
	case temperature < 20:
		fmt.Println("  Cold")
	case temperature < 30:
		fmt.Println("  Comfortable")
	default:
		fmt.Println("  Hot")
	}

	//  COMBINED EXAMPLE
	fmt.Println("\n--- 8. Combined Example ---")

	// Create map of Pet structs
	petDatabase := make(map[string]Pet)

	// Add pets to database
	petDatabase["P001"] = Pet{
		ID:           "P001",
		Name:         "Max",
		Species:      "Dog",
		Age:          3,
		IsVaccinated: true,
		Tags:         []string{"friendly", "trained"},
		Attributes:   map[string]string{"color": "Golden", "size": "Medium"},
	}

	petDatabase["P002"] = Pet{
		ID:           "P002",
		Name:         "Bella",
		Species:      "Cat",
		Age:          5,
		IsVaccinated: true,
		Tags:         []string{"calm", "indoor"},
		Attributes:   map[string]string{"color": "Black", "size": "Small"},
	}

	petDatabase["P003"] = Pet{
		ID:           "P003",
		Name:         "Charlie",
		Species:      "Dog",
		Age:          2,
		IsVaccinated: false,
		Tags:         []string{"playful", "young"},
		Attributes:   map[string]string{"color": "Brown", "size": "Large"},
	}

	// Loop through pets and apply control flow
	fmt.Println("Pet Database:")
	for id, pet := range petDatabase {
		fmt.Printf("\nPet ID: %s\n", id)
		fmt.Printf("  Name: %s, Species: %s, Age: %d\n", pet.Name, pet.Species, pet.Age)

		// Control flow for vaccination status
		if pet.IsVaccinated {
			fmt.Println("  Status: ✓ Vaccinated")
		} else {
			fmt.Println("  Status: ✗ Not Vaccinated - Needs attention!")
		}

		// Control flow for age category
		switch {
		case pet.Age < 2:
			fmt.Println("  Category: Puppy/Kitten")
		case pet.Age < 7:
			fmt.Println("  Category: Adult")
		default:
			fmt.Println("  Category: Senior")
		}

		// Loop through tags
		fmt.Print("  Tags: ")
		for i, tag := range pet.Tags {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(tag)
		}
		fmt.Println()

		// Loop through attributes
		fmt.Println("  Attributes:")
		for key, value := range pet.Attributes {
			fmt.Printf("    - %s: %s\n", key, value)
		}
	}
	fmt.Println("\n=== Demo Complete ===")
}

package main

import "fmt"

type Pet struct {
	Name    string
	Age     int
	Breed   string
	Adopted bool
}

func main() {
	var choice int

	for {
		fmt.Println("1. Variables, Values and Types")
		fmt.Println("2. Looping and Control Flow")
		fmt.Println("3. Array and Slice")
		fmt.Println("4. Map and Structs")
		fmt.Println("5. Exit")
		fmt.Print("Enter choice: ")
		fmt.Scan(&choice)

		switch choice {
		case 1:
			demonstrateVariables()
		case 2:
			demonstrateLooping()
		case 3:
			demonstrateArraySlice()
		case 4:
			demonstrateMapStruct()
		case 5:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func demonstrateVariables() {
	fmt.Println("\nVariables, Values and Types")

	var dogName string = "Buddy"
	age := 3
	weight := 12.5
	isVaccinated := true

	fmt.Printf("String (name): %s\n", dogName)
	fmt.Printf("Integer (age): %d years\n", age)
	fmt.Printf("Float (weight): %.1f kg\n", weight)
	fmt.Printf("Boolean (vaccinated): %t\n", isVaccinated)
}

func demonstrateLooping() {
	fmt.Println("\nLooping and Control Flow")

	fmt.Println("Pets available for adoption:")
	for i := 1; i <= 5; i++ {
		fmt.Printf("Pet ID: %d\n", i)
	}

	petAge := 2
	if petAge < 1 {
		fmt.Println("Category: Puppy")
	} else if petAge <= 7 {
		fmt.Println("Category: Adult")
	} else {
		fmt.Println("Category: Senior")
	}
}

func demonstrateArraySlice() {
	fmt.Println("\nArray and Slice")

	dogAges := [3]int{2, 5, 8}
	fmt.Printf("Dog ages (array): %v\n", dogAges)

	breeds := []string{"Labrador", "Beagle", "Indie"}
	fmt.Printf("Available breeds: %v\n", breeds)

	breeds = append(breeds, "Golden Retriever")
	fmt.Printf("After new rescue: %v\n", breeds)
	fmt.Printf("Total breeds: %d\n", len(breeds))
}

func demonstrateMapStruct() {
	fmt.Println("\nMap and Structs")

	// Map
	adoptionCount := map[string]int{
		"Labrador": 15,
		"Beagle":   8,
		"Indie":    22,
	}
	fmt.Println("Adoptions by breed:", adoptionCount)
	fmt.Printf("Labrador adoptions: %d\n", adoptionCount["Labrador"])

	// Struct
	pet := Pet{
		Name:    "Max",
		Age:     4,
		Breed:   "Labrador",
		Adopted: false,
	}
	fmt.Printf("Pet details: %+v\n", pet)
	fmt.Printf("Name: %s, Age: %d, Breed: %s, Adopted: %t\n", pet.Name, pet.Age, pet.Breed, pet.Adopted)
}

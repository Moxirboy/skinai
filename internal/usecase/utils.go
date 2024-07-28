package usecase

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"reflect"
	"testDeployment/internal/domain"
)

func Validator(str string) bool {
	return str == ""
}
func ValidatorAge(age int) bool {
	return age == 0
}
func calculatePercentage(data []bool) float64 {
	if len(data) == 0 {
		return 0.0
	}

	trueCount := 0
	for _, value := range data {
		trueCount += int(btoi(value))
	}

	percentage := float64(trueCount) / float64(len(data)) * 100.0
	return percentage
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
func calculateBMI(Weigh, Height string) float64 {
	weight, _ := strconv.ParseFloat(Weigh, 64)
	height, _ := strconv.ParseFloat(Height, 64)
	return weight / (height * height)*10000
}
func getRandomElement(slice []int) (int, error) {
	if len(slice) == 0 {
		return 0, fmt.Errorf("slice is empty")
	}

	// Initialize the random number generator with a seed based on current time
	rand.Seed(time.Now().UnixNano())

	// Get a random index
	randomIndex := rand.Intn(len(slice))

	// Get the random element
	randomElement := slice[randomIndex]

	return randomElement, nil
}
func getRandomExercises(exercises []int) ([]int, error) {
	var count int
	if len(exercises)<6{
		count=getRandomNumber(1, len(exercises))
	}else{
		count = getRandomNumber(3, 6)
	}
	
	if len(exercises) < count {
		return nil, fmt.Errorf("not enough exercises in the slice")
	}

	// Initialize the random number generator with a seed based on current time
	rand.Seed(time.Now().UnixNano())

	// Shuffle the original slice to ensure randomness
	shuffledExercises := make([]int, len(exercises))
	copy(shuffledExercises, exercises)
	rand.Shuffle(len(shuffledExercises), func(i, j int) {
		shuffledExercises[i], shuffledExercises[j] = shuffledExercises[j], shuffledExercises[i]
	})

	// Select the first 'count' elements from the shuffled slice
	selectedExercises := shuffledExercises[:count]

	return selectedExercises, nil
}

func getRandomNumber(min, max int) int {
	if min > max {
		min, max = max, min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}


func isAnyFieldEmpty(doctor domain.Doctor) bool {
	v := reflect.ValueOf(doctor)

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i).Interface()

		switch fieldValue := fieldValue.(type) {
		case int:
			if fieldValue == 0 {
				return true
			}
		case string:
			if fieldValue == "" {
				return true
			}
		// Add more cases for other types if necessary

		default:
			// Handle other types as needed
			return true
		}
	}

	return false
}
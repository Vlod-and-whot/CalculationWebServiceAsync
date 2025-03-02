package calculation

import (
	"errors"
	"os"
	"strconv"
	"time"
)

func GetOperationTime(op string) int {
	var envVar string
	switch op {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATIONS_MS"
	case "/":
		envVar = "TIME_DIVISIONS_MS"
	default:
		return 0
	}
	msStr := os.Getenv(envVar)
	if msStr == "" {
		return 1000
	}
	ms, err := strconv.Atoi(msStr)
	if err != nil {
		return 1000
	}
	return ms
}

func Compute(a, b float64, op string, delayMs int) (float64, error) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	default:
		return 0, errors.New("unsupported operation")
	}
}

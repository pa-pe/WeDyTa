package service

import "fmt"

func extractIsActive(record map[string]interface{}) bool {
	if isActive, exists := record["is_active"]; exists {
		return fmt.Sprint(isActive) == "1"
	}

	if isDisabled, exists := record["is_disabled"]; exists {
		return fmt.Sprint(isDisabled) == "0"
	}

	return true
}

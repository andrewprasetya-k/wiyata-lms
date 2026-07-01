package service

import "time"

func formatAPITime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func formatAPITimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatAPITime(*value)
	return &formatted
}

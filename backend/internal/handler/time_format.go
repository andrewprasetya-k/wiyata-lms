package handler

import "time"

func formatAPITime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

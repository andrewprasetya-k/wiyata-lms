package handler

import "time"

func formatUTCDateTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

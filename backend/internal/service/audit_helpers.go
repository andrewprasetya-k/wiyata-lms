package service

// strPtr is a small convenience for building the *string entityID/metadata
// values LogService.Log/LogBatch expect from a plain string.
func strPtr(s string) *string {
	return &s
}

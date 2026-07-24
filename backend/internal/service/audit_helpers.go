package service

// strPtr is a small convenience for building the *string entityID/metadata
// values LogService.Log/LogBatch expect from a plain string.
func strPtr(s string) *string {
	return &s
}

// nilIfEmpty is strPtr's counterpart for genuinely optional values (like an
// IP address that may not be available at a given call site) — an empty
// string becomes a real NULL column instead of a stored-but-meaningless "".
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

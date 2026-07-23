package service

import "golang.org/x/crypto/bcrypt"

// hashPassword centralizes bcrypt hashing so every password-setting call
// site (register, admin create/reset, invitation acceptance, super-admin
// bootstrap, CSV import default password) uses the same cost factor rather
// than each repeating bcrypt.GenerateFromPassword directly.
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// verifyPassword centralizes bcrypt comparison against a stored hash.
func verifyPassword(hashedPassword string, candidate string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidate))
}

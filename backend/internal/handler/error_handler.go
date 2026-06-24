package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HandleError centralizes error responses and masks internal details
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Print raw error to server logs for debugging
	fmt.Printf("[Error Log] %s\n", err.Error())

	// 1. Check for GORM Record Not Found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "The requested data was not found"})
		return
	}

	errStr := err.Error()

	// 2. Business Logic Errors (Custom validation from services)

	// Already exists errors
	if strings.Contains(errStr, "sudah terdaftar") || strings.Contains(errStr, "already exists") ||
		strings.Contains(errStr, "already assigned") {
		c.JSON(http.StatusConflict, gin.H{"error": "This data already exists or is already registered"})
		return
	}

	// Password errors
	if strings.Contains(errStr, "password lama salah") || strings.Contains(errStr, "incorrect password") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	// Deadline/submission errors
	if strings.Contains(errStr, "submission past due") || strings.Contains(errStr, "past deadline") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot submit past deadline"})
		return
	}

	if strings.Contains(errStr, "feed content is required") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feed content is required"})
		return
	}

	if strings.Contains(errStr, "comment content is required") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment content is required"})
		return
	}

	if strings.Contains(errStr, "student note content is required") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Isi catatan wajib diisi"})
		return
	}

	if strings.Contains(errStr, "student note content exceeds 10000 characters") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Catatan maksimal 10.000 karakter"})
		return
	}

	if strings.Contains(errStr, "unsupported comment source") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comments are only supported for feed posts in this MVP"})
		return
	}

	if strings.Contains(errStr, "assessment weights are required") ||
		strings.Contains(errStr, "assessment weight is required") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assessment weights are required"})
		return
	}

	if strings.Contains(errStr, "assessment weight must be between 0 and 100") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assessment weight must be between 0 and 100"})
		return
	}

	if strings.Contains(errStr, "duplicate assessment category in weights") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate assessment category in weights"})
		return
	}

	if strings.Contains(errStr, "total weight must be 100") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Total assessment weight must be 100"})
		return
	}

	if strings.Contains(errStr, "no weights configured for this subject") {
		c.JSON(http.StatusNotFound, gin.H{"error": "No weights configured for this subject"})
		return
	}

	if strings.Contains(errStr, "invalid media attachment") ||
		strings.Contains(errStr, "invalid assignment category") ||
		strings.Contains(errStr, "invalid assessment weight subject") ||
		strings.Contains(errStr, "invalid assessment weight category") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or non-existent data reference"})
		return
	}

	if strings.Contains(errStr, "forbidden:") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if strings.Contains(errStr, "failed to link media attachments") {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link media attachments"})
		return
	}

	// Cannot delete due to dependencies
	if strings.Contains(errStr, "teacher subject class assignment exists") {
		c.JSON(http.StatusConflict, gin.H{"error": "Teacher masih ditugaskan mengajar di kelas ini. Lepaskan penugasan mengajar terlebih dahulu."})
		return
	}

	if strings.Contains(errStr, "subject class has content") {
		c.JSON(http.StatusConflict, gin.H{"error": "Subject class masih memiliki materi atau tugas. Arsip/nonaktifkan subject class belum didukung."})
		return
	}

	if strings.Contains(errStr, "tidak bisa dihapus karena") || strings.Contains(errStr, "cannot be deleted") {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete this data because it is still being used by other records"})
		return
	}

	// 3. Database Constraint Errors (PostgreSQL patterns)

	// Foreign Key Violation
	if strings.Contains(errStr, "violates foreign key constraint") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or non-existent data reference"})
		return
	}

	// Unique Violation
	if strings.Contains(errStr, "duplicate key value violates unique constraint") {
		c.JSON(http.StatusConflict, gin.H{"error": "This data already exists in the system"})
		return
	}

	// Not Null Violation
	if strings.Contains(errStr, "violates not-null constraint") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required field is missing"})
		return
	}

	// Check Constraint Violation
	if strings.Contains(errStr, "violates check constraint") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data does not meet validation requirements"})
		return
	}

	// 4. Default Error (Internal Server Error)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
}

// HandleBindingError masks raw validation errors from Gin/Validator
func HandleBindingError(c *gin.Context, err error) {
	// Print raw binding error to server logs for debugging
	fmt.Printf("[Binding Error Log] %s\n", err.Error())

	errStr := err.Error()

	// Required field validation
	if strings.Contains(errStr, "failed on the 'required' tag") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Required fields are missing"})
		return
	}

	// UUID validation
	if strings.Contains(errStr, "failed on the 'uuid' tag") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format. Must be a valid UUID"})
		return
	}

	// Email validation
	if strings.Contains(errStr, "failed on the 'email' tag") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Enum validation (oneof)
	if strings.Contains(errStr, "failed on the 'oneof' tag") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value. Please check allowed values"})
		return
	}

	// Min/Max validation
	if strings.Contains(errStr, "failed on the 'min' tag") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Value is too short or too small"})
		return
	}
	if strings.Contains(errStr, "failed on the 'max' tag") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Value is too long or too large"})
		return
	}

	// Dive validation (array elements)
	if strings.Contains(errStr, "failed on the 'dive' tag") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more array elements are invalid"})
		return
	}

	// Type mismatch errors (e.g. sending string instead of number)
	if strings.Contains(errStr, "unmarshal") || strings.Contains(errStr, "type mismatch") ||
		strings.Contains(errStr, "cannot unmarshal") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data type mismatch. Please check your input values"})
		return
	}

	// JSON syntax errors
	if strings.Contains(errStr, "invalid character") || strings.Contains(errStr, "unexpected end of JSON") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Default binding error
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data format. Please check your request"})
}

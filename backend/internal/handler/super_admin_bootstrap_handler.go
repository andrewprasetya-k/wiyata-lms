package handler

import (
	"backend/internal/domain"
	"backend/internal/dto"
	"backend/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SuperAdminBootstrapHandler struct {
	service service.SuperAdminBootstrapService
}

func NewSuperAdminBootstrapHandler(service service.SuperAdminBootstrapService) *SuperAdminBootstrapHandler {
	return &SuperAdminBootstrapHandler{service: service}
}

func (h *SuperAdminBootstrapHandler) BootstrapSchool(c *gin.Context) {
	var input dto.SchoolBootstrapRequestDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		HandleBindingError(c, err)
		return
	}

	actor := buildActorContext(c, domain.LogScopePlatform)
	response, err := h.service.BootstrapSchool(actor, input)
	if err != nil {
		handleSchoolBootstrapError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func handleSchoolBootstrapError(c *gin.Context, err error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "bootstrap duplicate school code"):
		c.JSON(http.StatusConflict, gin.H{"error": "Kode sekolah sudah digunakan"})
	case strings.Contains(errStr, "bootstrap duplicate school email"):
		c.JSON(http.StatusConflict, gin.H{"error": "Email sekolah sudah digunakan"})
	case strings.Contains(errStr, "bootstrap duplicate school phone"):
		c.JSON(http.StatusConflict, gin.H{"error": "Nomor telepon sekolah sudah digunakan"})
	case strings.Contains(errStr, "bootstrap duplicate user email"):
		c.JSON(http.StatusConflict, gin.H{"error": "Email admin sekolah sudah digunakan"})
	case strings.Contains(errStr, "bootstrap existing admin user not found"):
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun admin sekolah tidak ditemukan atau tidak aktif"})
	case strings.Contains(errStr, "bootstrap admin role not found"):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role admin belum tersedia di sistem"})
	case strings.Contains(errStr, "bootstrap invalid admin user mode"):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mode adminUser harus new atau existing"})
	case strings.Contains(errStr, "bootstrap new admin user fields are required"):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama, email, dan password admin sekolah wajib diisi"})
	case strings.Contains(errStr, "bootstrap invalid admin user email"):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email admin sekolah tidak valid"})
	case strings.Contains(errStr, "bootstrap admin user password too short"):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password admin sekolah minimal 6 karakter"})
	case strings.Contains(errStr, "bootstrap existing admin user id is required"):
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID admin sekolah wajib diisi"})
	case strings.Contains(errStr, "bootstrap invalid existing admin user id"):
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID admin sekolah tidak valid"})
	case strings.Contains(errStr, "bootstrap school code generation failed"):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kode sekolah otomatis belum bisa dibuat"})
	default:
		HandleError(c, err)
	}
}

package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

type EmailService interface {
	SendSchoolAdminInvitation(toEmail string, schoolName string, acceptURL string) error
	SendSchoolMemberInvitation(toEmail string, schoolName string, role string, acceptURL string) error
	SendSchoolMemberAccountCreated(toEmail string, schoolName string, role string) error
	SendSchoolMemberAddedToSchool(toEmail string, schoolName string, role string) error
	SendEmailVerification(toEmail string, fullName string, verifyURL string) error
}

type noopEmailService struct{}

func NewEmailServiceFromEnv() EmailService {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("SMTP_ENABLED")), "true")
	config := smtpEmailConfig{
		Host:      strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:      strings.TrimSpace(os.Getenv("SMTP_PORT")),
		Username:  strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password:  os.Getenv("SMTP_PASSWORD"),
		FromEmail: strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		FromName:  strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
	}

	if !enabled || !config.isComplete() {
		return noopEmailService{}
	}
	if config.FromName == "" {
		config.FromName = "Wiyata"
	}
	if config.Port == "" {
		config.Port = "587"
	}

	return &smtpEmailService{config: config}
}

func (noopEmailService) SendSchoolAdminInvitation(string, string, string) error {
	return nil
}

func (noopEmailService) SendSchoolMemberInvitation(string, string, string, string) error {
	return nil
}

func (noopEmailService) SendSchoolMemberAccountCreated(string, string, string) error {
	return nil
}

func (noopEmailService) SendSchoolMemberAddedToSchool(string, string, string) error {
	return nil
}

func (noopEmailService) SendEmailVerification(toEmail string, fullName string, verifyURL string) error {
	fmt.Println("EMAIL VERIFICATION")
	fmt.Println()
	fmt.Printf("User : %s <%s>\n", fullName, toEmail)
	fmt.Println()
	fmt.Println("Verify URL:")
	fmt.Println(verifyURL)
	return nil
}

type smtpEmailConfig struct {
	Host      string
	Port      string
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

func (c smtpEmailConfig) isComplete() bool {
	return c.Host != "" &&
		c.Username != "" &&
		c.Password != "" &&
		c.FromEmail != ""
}

type smtpEmailService struct {
	config smtpEmailConfig
}

func (s *smtpEmailService) SendSchoolAdminInvitation(toEmail string, schoolName string, acceptURL string) error {
	toEmail = strings.TrimSpace(toEmail)
	schoolName = strings.TrimSpace(schoolName)
	acceptURL = strings.TrimSpace(acceptURL)
	if toEmail == "" || schoolName == "" || acceptURL == "" {
		return fmt.Errorf("email invitation fields are required")
	}

	subject := "Undangan Admin Sekolah Wiyata"
	body := fmt.Sprintf(`Halo,

Anda diundang menjadi Admin Sekolah untuk %s di Wiyata.

Gunakan link berikut untuk menerima undangan dan membuat password:
%s

Jika Anda tidak meminta akses ini, abaikan email ini.

Salam,
Wiyata
`, schoolName, acceptURL)

	return s.sendPlainText(toEmail, subject, body)
}

func (s *smtpEmailService) SendSchoolMemberInvitation(toEmail string, schoolName string, role string, acceptURL string) error {
	toEmail = strings.TrimSpace(toEmail)
	schoolName = strings.TrimSpace(schoolName)
	role = strings.TrimSpace(strings.ToLower(role))
	acceptURL = strings.TrimSpace(acceptURL)
	if toEmail == "" || schoolName == "" || role == "" || acceptURL == "" {
		return fmt.Errorf("email invitation fields are required")
	}

	roleLabel := "Warga Sekolah"
	if role == "teacher" {
		roleLabel = "Guru"
	}
	if role == "student" {
		roleLabel = "Siswa"
	}

	subject := "Undangan Bergabung di Wiyata"
	body := fmt.Sprintf(`Halo,

Anda diundang menjadi %s di %s.

Gunakan link berikut untuk menerima undangan dan membuat password:
%s

Link undangan ini memiliki masa berlaku terbatas. Jika Anda tidak meminta akses ini, abaikan email ini.

Salam,
Wiyata
`, roleLabel, schoolName, acceptURL)

	return s.sendPlainText(toEmail, subject, body)
}

func (s *smtpEmailService) SendSchoolMemberAccountCreated(toEmail string, schoolName string, role string) error {
	toEmail = strings.TrimSpace(toEmail)
	schoolName = strings.TrimSpace(schoolName)
	roleLabel := schoolMemberRoleLabel(role)
	if toEmail == "" || schoolName == "" {
		return fmt.Errorf("email account-created fields are required")
	}

	subject := "Akun Wiyata Anda Sudah Dibuat"
	body := fmt.Sprintf(`Halo,

Akun Wiyata Anda sudah dibuat dan ditambahkan sebagai %s di %s.

Password awal diberikan langsung oleh admin/sekolah. Demi keamanan, Wiyata tidak mengirim password melalui email.

Silakan login menggunakan email ini dan password awal yang diberikan oleh admin/sekolah.

Salam,
Wiyata
`, roleLabel, schoolName)

	return s.sendPlainText(toEmail, subject, body)
}

func (s *smtpEmailService) SendSchoolMemberAddedToSchool(toEmail string, schoolName string, role string) error {
	toEmail = strings.TrimSpace(toEmail)
	schoolName = strings.TrimSpace(schoolName)
	roleLabel := schoolMemberRoleLabel(role)
	if toEmail == "" || schoolName == "" {
		return fmt.Errorf("email added-to-school fields are required")
	}

	subject := "Anda Ditambahkan ke Sekolah di Wiyata"
	body := fmt.Sprintf(`Halo,

Akun Wiyata Anda sudah ditambahkan sebagai %s di %s.

Password akun Anda tidak diubah. Silakan login menggunakan password akun Wiyata yang sudah ada.

Salam,
Wiyata
`, roleLabel, schoolName)

	return s.sendPlainText(toEmail, subject, body)
}

func (s *smtpEmailService) SendEmailVerification(toEmail string, fullName string, verifyURL string) error {
	toEmail = strings.TrimSpace(toEmail)
	fullName = strings.TrimSpace(fullName)
	verifyURL = strings.TrimSpace(verifyURL)
	if toEmail == "" || verifyURL == "" {
		return fmt.Errorf("email verification fields are required")
	}

	subject := "Verifikasi Email Wiyata"
	body := fmt.Sprintf(`Halo %s,

Terima kasih sudah mendaftar di Wiyata. Verifikasi email Anda untuk mulai menggunakan Wiyata sepenuhnya.

Gunakan link berikut untuk memverifikasi email:
%s

Link ini berlaku selama 24 jam. Jika Anda tidak mendaftar di Wiyata, abaikan email ini.

Salam,
Wiyata
`, fullName, verifyURL)

	return s.sendPlainText(toEmail, subject, body)
}

func schoolMemberRoleLabel(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "teacher":
		return "Guru"
	case "student":
		return "Siswa"
	case "admin":
		return "Admin Sekolah"
	default:
		return "Warga Sekolah"
	}
}

func (s *smtpEmailService) sendPlainText(toEmail string, subject string, body string) error {
	message := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", s.config.FromName, s.config.FromEmail),
		fmt.Sprintf("To: %s", toEmail),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	address := net.JoinHostPort(s.config.Host, s.config.Port)
	if s.config.Port == "587" {
		return s.sendWithStartTLS(address, toEmail, []byte(message))
	}

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	return smtp.SendMail(address, auth, s.config.FromEmail, []string{toEmail}, []byte(message))
}

func (s *smtpEmailService) sendWithStartTLS(address string, toEmail string, message []byte) error {
	client, err := smtp.Dial(address)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("smtp server does not support STARTTLS")
	}
	if err := client.StartTLS(&tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12,
	}); err != nil {
		return err
	}

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(s.config.FromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

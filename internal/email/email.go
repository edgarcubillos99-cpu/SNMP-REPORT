package email

import (
	"net/smtp"
	"os"
)

// Enviar reporte por correo electrónico
func SendReport(path string, to string) error {

	// Leer archivo de reporte
	bodyFile, _ := os.ReadFile(path)
	msg := "Subject: Reporte SNMP\n\n" + string(bodyFile)

	// Configurar autenticación SMTP
	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASS"),
		os.Getenv("SMTP_HOST"),
	)

	// Enviar correo
	server := os.Getenv("SMTP_HOST") + ":" + os.Getenv("SMTP_PORT")

	// Enviar el correo a la dirección especificada
	return smtp.SendMail(
		server,
		auth,
		os.Getenv("SMTP_USER"),
		[]string{to},
		[]byte(msg),
	)
}

// aqui se definen las funciones para enviar correos electrónicos con reportes
package email

import (
	"fmt"
	"net/smtp"
	"os"
)

// Enviar reporte por correo electrónico
func SendReport(path string, to string) error {

	// leer archivo de reporte
	bodyFile, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("no se pudo leer el reporte %s: %w", path, err)
	}

	// construir mensaje
	msg := "Subject: Reporte SNMP\r\n" +
		"Content-Type: application/json; charset=\"utf-8\"\r\n" +
		"\r\n" +
		string(bodyFile)

	// configurar autenticación SMTP
	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASS"),
		os.Getenv("SMTP_HOST"),
	)

	server := os.Getenv("SMTP_HOST") + ":" + os.Getenv("SMTP_PORT")

	// enviar correo
	return smtp.SendMail(
		server,
		auth,
		os.Getenv("SMTP_USER"),
		[]string{to},
		[]byte(msg),
	)
}

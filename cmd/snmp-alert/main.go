package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"snmp-alert/internal/config"
	"snmp-alert/internal/email"
	"snmp-alert/internal/report"
	"snmp-alert/internal/snmp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {

	// Cargar variables de entorno
	godotenv.Load()

	// Cargar configuración de agentes SNMP
	cfg, err := config.LoadConfig("configs/agents.json")
	if err != nil {
		log.Fatal(err)
	}

	// Canal desde traps y polling
	reportChan := make(chan report.ReportItem, 100)

	// Lista protegida por mutex
	var (
		reportList []report.ReportItem
		mu         sync.Mutex
	)

	// Mapa para evitar alertas repetidas por IP + OID
	lastAlert := make(map[string]time.Time)
	var lastAlertMu sync.Mutex

	// ================================
	// 🔧 Leer ALERT_REPEAT_MINUTES del .env
	// ================================
	repeatMinutesStr := os.Getenv("ALERT_REPEAT_MINUTES")
	repeatMinutes, err := strconv.Atoi(repeatMinutesStr)
	if err != nil || repeatMinutes <= 0 {
		repeatMinutes = 30 // valor por defecto
	}

	minInterval := time.Duration(repeatMinutes) * time.Minute
	fmt.Println("⏱ Tiempo mínimo entre alertas repetidas:", minInterval)

	// Iniciar servidor de traps
	go snmp.StartTrapServer(reportChan)

	// Iniciar polling cada 10 segundos
	go snmp.StartPolling(cfg, reportChan, 10)

	fmt.Println("🟢 Sistema SNMP iniciado (Polling + Traps)")

	// ========== SCHEDULER AUTOMÁTICO ==========
	reportInterval := 10 // segundos

	go func() {
		for {
			time.Sleep(time.Duration(reportInterval) * time.Second)

			mu.Lock()
			if len(reportList) == 0 {
				mu.Unlock()
				fmt.Println("⏳ No hay alertas nuevas para reportar.")
				continue
			}

			// Copiar y limpiar buffer de alertas
			toSend := make([]report.ReportItem, len(reportList))
			copy(toSend, reportList)
			reportList = []report.ReportItem{}
			mu.Unlock()

			// Guardar reporte
			if err := report.Save(toSend, "output/report.json"); err != nil {
				fmt.Println("❌ Error guardando reporte:", err)
				continue
			}

			fmt.Println("📤 Enviando reporte automático con", len(toSend), "alertas...")

			// Enviar correo
			if err := email.SendReport("output/report.json", os.Getenv("ALERT_EMAIL")); err != nil {
				fmt.Println("❌ Error al enviar correo:", err)
			} else {
				fmt.Println("✅ Reporte enviado exitosamente.")
			}
		}
	}()

	// Señal de salida
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {

		case r := <-reportChan:
			// Crear clave única (IP + OID)
			key := r.IP + "|" + r.OID

			lastAlertMu.Lock()
			lastTime, exists := lastAlert[key]
			now := time.Now()

			// Si existe y no han pasado X minutos → ignorar alerta
			if exists && now.Sub(lastTime) < minInterval {
				lastAlertMu.Unlock()
				fmt.Printf("⏳ Alerta repetida omitida (%s - %s)\n", r.IP, r.OID)
				continue
			}

			// Registrar nuevo timestamp para esta alerta
			lastAlert[key] = now
			lastAlertMu.Unlock()

			// Agregar alerta al buffer para el reporte
			mu.Lock()
			reportList = append(reportList, r)
			mu.Unlock()

			fmt.Println("⚠ Alerta nueva registrada:", r)

		case <-exitChan:
			fmt.Println("⛔ Deteniendo servicio... SIN enviar reporte final.")
			return
		}
	}
}

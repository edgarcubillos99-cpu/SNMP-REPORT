// aqui se definen las funciones para realizar el polling SNMP y evaluar las reglas configuradas
package snmp

import (
	"fmt"
	"time"

	"snmp-alert/internal/config"
	"snmp-alert/internal/report"
	"snmp-alert/internal/rules"
)

func StartPolling(cfg *config.Config, reportChan chan<- report.ReportItem, interval int) {

	// Bucle infinito de polling
	for {
		fmt.Println("🔎 Ejecutando SNMP GET polling...")

		for _, agent := range cfg.Agents {
			values, err := Query(agent)
			if err != nil {
				fmt.Println("Error en polling:", err)
				continue
			}

			// Evaluar las reglas de cada OID
			for _, rule := range agent.OIDs {

				val, ok := values[rule.OID]
				if !ok {
					fmt.Println("⚠ No se recibió valor para:", rule.OID)
					continue
				}

				// ==========================
				// Selección automática:
				// - Si rule.Rule es between o outside → usa ValueMin + ValueMax
				// - Si rule.Rule es simple → usa Value
				// ==========================

				var isRisk bool

				switch rule.Rule {

				case "between", "outside":
					// Se requieren 2 valores (min y max)
					isRisk = rules.IsRisk(
						val,
						rule.Rule,
						rule.ValueMin,
						rule.ValueMax,
					)

				default:
					// Reglas simples: >, <, >=, <=, ==, !=
					isRisk = rules.IsRisk(
						val,
						rule.Rule,
						rule.Value,
					)
				}

				// Si hay riesgo, enviar alerta al canal
				if isRisk {
					reportChan <- report.ReportItem{
						IP:    agent.IP,
						OID:   rule.OID,
						Value: fmt.Sprintf("%v", val),
						Alert: "CRÍTICO (GET)",
					}
				}
			}
		}

		// Esperar el intervalo antes del siguiente polling
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

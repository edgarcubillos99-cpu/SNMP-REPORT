// aqui se definen las funciones para realizar el polling SNMP y evaluar las reglas configuradas
package snmp

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"snmp-alert/internal/config"
	"snmp-alert/internal/report"
	"snmp-alert/internal/rules"
)

func StartPolling(cfg *config.Config, reportChan chan<- report.ReportItem, interval int) {

	// definir número de workers
	const workerCount = 5 // o len(cfg.Agents) si queremos 1 worker por agente
	jobs := make(chan config.Agent)

	var wg sync.WaitGroup

	// Workers
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for agent := range jobs {
				values, err := Query(agent)
				if err != nil {
					fmt.Println("Error en polling (worker", id, "):", err)
					wg.Done()
					continue
				}

				// evaluar reglas
				for _, rule := range agent.OIDs {

					val, ok := values[rule.OID]
					if !ok {
						fmt.Println("⚠ No se recibió valor para:", rule.OID, "en", agent.IP)
						continue
					}

					// determinar tipo de regla
					ruleName := strings.ToLower(strings.TrimSpace(rule.Rule))
					var isRisk bool

					// evaluar según tipo de regla
					switch ruleName {
					case "between", "outside": // reglas de rango
						isRisk = rules.IsRisk(
							val,
							ruleName,
							rule.ValueMin,
							rule.ValueMax,
						)
					default: // reglas de valor simple
						isRisk = rules.IsRisk(
							val,
							ruleName,
							rule.Value,
						)
					}

					// si hay riesgo, enviar reporte
					if isRisk {
						reportChan <- report.ReportItem{
							IP:    agent.IP,
							OID:   rule.OID,
							Value: fmt.Sprintf("%v", val),
							Alert: "CRÍTICO (GET)",
						}
					}
				}

				wg.Done()
			}
		}(i + 1)
	}

	// Bucle de scheduling
	for {
		fmt.Println("🔎 Ejecutando SNMP GET polling (worker pool)...")

		// enviar trabajos a los workers
		for _, agent := range cfg.Agents {
			wg.Add(1)
			jobs <- agent
		}

		// Esperar a que termine la ronda
		wg.Wait()

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

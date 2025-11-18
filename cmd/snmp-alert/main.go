package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/joho/godotenv"

	"snmp-alert/internal/config"
	"snmp-alert/internal/email"
	"snmp-alert/internal/report"
	"snmp-alert/internal/rules"
	"snmp-alert/internal/snmp"
)

func main() {

	// Cargar .env
	godotenv.Load()

	// Cargar configuración de agentes SNMP
	cfg, err := config.LoadConfig("configs/agents.json")
	if err != nil {
		log.Fatal(err)
	}

	// Consultar agentes SNMP en paralelo
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Resultado de los reportes
	var rep []report.ReportItem

	// Iterar sobre los agentes
	for _, agent := range cfg.Agents {
		wg.Add(1)

		go func(a config.Agent) {
			defer wg.Done()

			// Consultar agente SNMP
			values, err := snmp.Query(a)
			if err != nil {
				return
			}

			// Evaluar reglas para cada OID
			for _, r := range a.OIDs {
				v := values[r.OID]

				if rules.IsRisk(v, r.Rule, r.Value) {
					mu.Lock()
					rep = append(rep, report.ReportItem{
						IP:    a.IP,
						OID:   r.OID,
						Value: fmt.Sprintf("%v", v),
						Alert: "CRÍTICO",
					})
					mu.Unlock()
				}
			}

		}(agent)
	}

	// Esperar a que terminen todas las consultas
	wg.Wait()

	report.Save(rep, "output/report.json")

	// Usar correo desde .env
	email.SendReport("output/report.json", "destino@empresa.com")

	fmt.Println("Reporte enviado!")
}

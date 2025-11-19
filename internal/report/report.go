// aqui se definen las estructuras y funciones para generar reportes en formato JSON
package report

import (
	"encoding/json"
	"os"
)

// Definición de un ítem de reporte
type ReportItem struct {
	IP    string `json:"ip"`
	OID   string `json:"oid"`
	Value string `json:"value"`
	Alert string `json:"alert"`
}

// Guardar reporte en un archivo JSON
func Save(report []ReportItem, path string) error {
	data, _ := json.MarshalIndent(report, "", "  ")
	return os.WriteFile(path, data, 0644)
}

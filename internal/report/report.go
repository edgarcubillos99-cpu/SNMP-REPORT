// aqui se definen las estructuras y funciones para generar reportes en formato JSON
package report

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	// crear directorios si no existen
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, data, 0644)
}

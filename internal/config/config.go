package config

import (
	"encoding/json"
	"os"
)

// Definiciones de estructuras de configuración
type OIDRule struct {
	OID   string  `json:"oid"`
	Rule  string  `json:"rule"`
	Value float64 `json:"value"`
}

// Definición de un agente SNMP
type Agent struct {
	IP          string    `json:"ip"`
	SNMPVersion string    `json:"snmp_version"` // v2 o v3
	Community   string    `json:"community,omitempty"`
	OIDs        []OIDRule `json:"oids"`
}

// archivo json de configuración
type Config struct {
	Agents []Agent `json:"agents"`
}

// Cargar configuración desde el archivo JSON
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// convertir JSON a estructura Config
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	return &cfg, err
}

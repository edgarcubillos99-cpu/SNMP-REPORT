// aqui se definen las funciones para consultar agentes SNMP
package snmp

import (
	"fmt"
	"os"
	"strings"
	"time"

	"snmp-alert/internal/config"

	"github.com/gosnmp/gosnmp"
)

// Consultar un agente SNMP y retornar los valores de los OIDs especificados
func Query(agent config.Agent) (map[string]float64, error) {

	var g *gosnmp.GoSNMP

	// Puerto por defecto
	port := uint16(161)
	if agent.Port != 0 {
		port = agent.Port
	}

	// Normalizar versión
	version := strings.ToLower(strings.TrimSpace(agent.SNMPVersion))

	// Configurar cliente SNMP según versión
	switch version {
	case "v2", "v2c":
		g = &gosnmp.GoSNMP{
			Target:    agent.IP,
			Community: agent.Community,
			Port:      port,
			Version:   gosnmp.Version2c,
			Timeout:   2 * time.Second,
		}

	case "v3":
		authProto := gosnmp.NoAuth
		privProto := gosnmp.NoPriv

		switch os.Getenv("SNMPV3_AUTH_PROTOCOL") {
		case "MD5":
			authProto = gosnmp.MD5
		case "SHA":
			authProto = gosnmp.SHA
		}

		switch os.Getenv("SNMPV3_PRIV_PROTOCOL") {
		case "AES":
			privProto = gosnmp.AES
		case "DES":
			privProto = gosnmp.DES
		}

		g = &gosnmp.GoSNMP{
			Target:        agent.IP,
			Port:          port,
			Version:       gosnmp.Version3,
			Timeout:       2 * time.Second,
			SecurityModel: gosnmp.UserSecurityModel,
			MsgFlags:      gosnmp.AuthPriv,
			SecurityParameters: &gosnmp.UsmSecurityParameters{
				UserName:                 os.Getenv("SNMPV3_USER"),
				AuthenticationProtocol:   authProto,
				AuthenticationPassphrase: os.Getenv("SNMPV3_AUTH_PWD"),
				PrivacyProtocol:          privProto,
				PrivacyPassphrase:        os.Getenv("SNMPV3_PRIV_PWD"),
			},
		}

	// si no es ninguna versión soportada
	default:
		return nil, fmt.Errorf("versión SNMP no soportada: %s", agent.SNMPVersion)
	}

	// conectar al agente SNMP
	if err := g.Connect(); err != nil {
		return nil, err
	}
	defer g.Conn.Close()

	// mapa para almacenar resultados
	results := make(map[string]float64)

	// consultar cada OID
	for _, o := range agent.OIDs {
		pdu, err := g.Get([]string{o.OID})
		if err != nil {
			fmt.Println("Error GET en", agent.IP, o.OID, err)
			continue
		}

		// verificar que el PDU no esté vacío
		if pdu == nil || len(pdu.Variables) == 0 {
			fmt.Println("Empty PDU en", agent.IP, o.OID)
			continue
		}

		// convertir el valor del OID a float64
		bi := gosnmp.ToBigInt(pdu.Variables[0].Value)
		if bi == nil {
			fmt.Println("Error conversión (ToBigInt) en", agent.IP, o.OID)
			continue
		}

		// almacenar el valor en el mapa de resultados
		v, _ := bi.Float64()
		results[o.OID] = v
	}

	return results, nil
}

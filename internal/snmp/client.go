// aqui se definen las funciones para consultar agentes SNMP
package snmp

import (
	"fmt"
	"os"
	"time"

	"snmp-alert/internal/config"

	"github.com/gosnmp/gosnmp"
)

// Consultar un agente SNMP y retornar los valores de los OIDs especificados
func Query(agent config.Agent) (map[string]float64, error) {

	// Configuración del cliente SNMP
	var g *gosnmp.GoSNMP

	// -------------------------
	// SNMP v2c
	// -------------------------
	if agent.SNMPVersion == "v2" {
		g = &gosnmp.GoSNMP{
			Target:    agent.IP,
			Community: agent.Community,
			Port:      161,
			Version:   gosnmp.Version2c,
			Timeout:   2 * time.Second,
		}
	}

	// -------------------------
	// SNMP v3 - con variables de entorno
	// -------------------------
	if agent.SNMPVersion == "v3" {

		authProto := gosnmp.NoAuth
		privProto := gosnmp.NoPriv

		// PROTOCOLO DE AUTENTICACIÓN
		switch os.Getenv("SNMPV3_AUTH_PROTOCOL") {
		case "MD5":
			authProto = gosnmp.MD5
		case "SHA":
			authProto = gosnmp.SHA
		}

		// PROTOCOLO DE PRIVACIDAD
		switch os.Getenv("SNMPV3_PRIV_PROTOCOL") {
		case "AES":
			privProto = gosnmp.AES
		case "DES":
			privProto = gosnmp.DES
		}

		// Configuración del cliente SNMP v3
		g = &gosnmp.GoSNMP{
			Target:        agent.IP,
			Port:          161,
			Version:       gosnmp.Version3,
			Timeout:       2 * time.Second,
			SecurityModel: gosnmp.UserSecurityModel,

			MsgFlags: gosnmp.AuthPriv, // se puede ajustar según necesidad

			SecurityParameters: &gosnmp.UsmSecurityParameters{
				UserName:                 os.Getenv("SNMPV3_USER"),
				AuthenticationProtocol:   authProto,
				AuthenticationPassphrase: os.Getenv("SNMPV3_AUTH_PWD"),
				PrivacyProtocol:          privProto,
				PrivacyPassphrase:        os.Getenv("SNMPV3_PRIV_PWD"),
			},
		}
	}

	// -------------------------
	// CONEXIÓN SNMP
	// -------------------------
	if err := g.Connect(); err != nil {
		return nil, err
	}
	defer g.Conn.Close()

	results := make(map[string]float64)

	// Iterar sobre los OIDs a consultar con GET {o.OID}
	for _, o := range agent.OIDs {
		pdu, err := g.Get([]string{o.OID})
		if err != nil {
			fmt.Println("Error GET en", agent.IP, o.OID, err)
			continue
		}

		// Conversión segura del valor obtenido
		if pdu == nil || len(pdu.Variables) == 0 {
			fmt.Println("Empty PDU en", agent.IP, o.OID)
			continue
		}

		// Convertir a BigInt primero
		bi := gosnmp.ToBigInt(pdu.Variables[0].Value)
		if bi == nil {
			fmt.Println("Error conversión (ToBigInt) en", agent.IP, o.OID)
			continue
		}

		// Convertir a float64
		v, _ := bi.Float64()

		results[o.OID] = v
	}

	return results, nil
}

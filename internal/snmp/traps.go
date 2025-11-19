package snmp

import (
	"fmt"
	"log"
	"net"
	"snmp-alert/internal/report"

	"github.com/gosnmp/gosnmp"
)

func StartTrapServer(reportChan chan<- report.ReportItem) {

	trapListener := gosnmp.NewTrapListener()
	trapListener.Params = gosnmp.Default

	trapListener.OnNewTrap = func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
		fmt.Println("📩 Trap recibido desde:", addr.IP.String())

		for _, v := range packet.Variables {

			ri := report.ReportItem{
				IP:    addr.IP.String(),
				OID:   v.Name,
				Value: fmt.Sprintf("%v", v.Value),
				Alert: "TRAP",
			}

			// Enviar a canal de reportes
			reportChan <- ri

			fmt.Printf("OID: %s → %v\n", v.Name, v.Value)
		}
	}

	fmt.Println("📡 Escuchando SNMP TRAPS en UDP/162...")

	err := trapListener.Listen("0.0.0.0:162")
	if err != nil {
		log.Fatalf("Error al iniciar Trap Server: %v", err)
	}
}

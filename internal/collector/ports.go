package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

type PortCollector struct {
	wellKnownPorts map[uint16]string
}

func NewPortCollector(wellKnown map[uint16]string) *PortCollector {
	if wellKnown == nil {
		wellKnown = defaultWellKnownPorts()
	}
	return &PortCollector{wellKnownPorts: wellKnown}
}

func (p *PortCollector) SetWellKnown(wellKnown map[uint16]string) {
	p.wellKnownPorts = wellKnown
}

func (p *PortCollector) Collect(lookupPID func(uint32) string) []PortBinding {
	entries := linuxproc.ParsePorts()
	bindings := make([]PortBinding, 0, len(entries))

	for _, e := range entries {
		proto := e.Protocol
		if proto == "tcp6" {
			proto = "tcp"
		} else if proto == "udp6" {
			proto = "udp"
		}

		label := p.wellKnownPorts[uint16(e.LocalPort)]

		bindings = append(bindings, PortBinding{
			Protocol:   proto,
			LocalAddr:  e.LocalAddr,
			LocalPort:  uint16(e.LocalPort),
			RemoteAddr: e.RemoteAddr,
			RemotePort: uint16(e.RemotePort),
			State:      e.State,
			StateCode:  e.StateCode,
			PID:        e.PID,
			Process:    e.ProcessName,
			Label:      label,
			Since:      0,
		})
	}

	return bindings
}

func defaultWellKnownPorts() map[uint16]string {
	return map[uint16]string{
		22:    "SSH",
		80:    "HTTP",
		443:   "HTTPS",
		3000:  "Dev Server",
		5432:  "PostgreSQL",
		6379:  "Redis",
		8080:  "HTTP Alt",
		9090:  "Prometheus",
		27017: "MongoDB",
	}
}
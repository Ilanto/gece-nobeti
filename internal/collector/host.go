package collector

import "github.com/burak/linux-dashboard/internal/linuxproc"

// HostCollector reads static system identity info.
type HostCollector struct{}

func NewHostCollector() *HostCollector { return &HostCollector{} }

func (h *HostCollector) Collect() HostMetrics {
	info := linuxproc.CollectHost()
	return HostMetrics{
		Hostname:      info.Hostname,
		KernelVersion: info.KernelVersion,
		Arch:          info.Arch,
		OS:            info.OS,
		UptimeSeconds: info.UptimeSeconds,
		User:          info.User,
		UID:           info.UID,
		Shell:         info.Shell,
		TTY:           info.TTY,
	}
}

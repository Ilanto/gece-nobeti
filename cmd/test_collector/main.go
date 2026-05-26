package main

import (
	"fmt"

	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/event"
)

func main() {
	fmt.Println("=== Collector Test ===")

	emit := event.NewEmitter()
	col := collector.NewManager(emit)

	// Collect once manually
	col.Collect()

	snap := col.LatestSnapshot()
	if snap == nil {
		fmt.Println("ERROR: LatestSnapshot() is nil after Collect()")
		return
	}

	fmt.Printf("CPU Total: %.2f%%\n", snap.CPU.TotalPercent)
	fmt.Printf("Memory Used: %d MB / %d MB (%.1f%%)\n",
		snap.Memory.UsedPhys/1024/1024,
		snap.Memory.TotalPhys/1024/1024,
		snap.Memory.UsedPercent)
	fmt.Printf("Processes: %d\n", len(snap.Processes))
	if len(snap.Network.Interfaces) > 0 {
		fmt.Printf("Network RX: %d bytes/s\n", snap.Network.Interfaces[0].InBPS)
	}
	fmt.Printf("Timestamp: %d\n", snap.Timestamp)
	fmt.Printf("Disk count: %d\n", len(snap.Disk.Drives))

	col.Stop()
	fmt.Println("=== Test OK ===")
}
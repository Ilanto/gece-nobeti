// Package linuxproc provides low-level access to Linux's /proc filesystem.
package linuxproc

import (
	"bufio"
	"encoding/hex"
	"net"
	"os"
	"strconv"
	"strings"
)

// PortEntry holds TCP/UDP port information.
type PortEntry struct {
	Protocol    string
	LocalAddr   string
	LocalPort   uint32
	RemoteAddr  string
	RemotePort  uint32
	State       string
	StateCode   uint32
	PID         uint32
	ProcessName string
}

// ParsePorts reads /proc/net/tcp, /proc/net/udp and returns all port bindings.
func ParsePorts() []PortEntry {
	var entries []PortEntry

	// TCP
	tcpEntries := parseProcNet("tcp")
	for _, e := range tcpEntries {
		e.Protocol = "tcp"
		entries = append(entries, e)
	}

	// TCP6
	tcp6Entries := parseProcNet("tcp6")
	for _, e := range tcp6Entries {
		e.Protocol = "tcp6"
		entries = append(entries, e)
	}

	// UDP
	udpEntries := parseProcNet("udp")
	for _, e := range udpEntries {
		e.Protocol = "udp"
		entries = append(entries, e)
	}

	// UDP6
	udp6Entries := parseProcNet("udp6")
	for _, e := range udp6Entries {
		e.Protocol = "udp6"
		entries = append(entries, e)
	}

	return entries
}

func parseProcNet(proto string) []PortEntry {
	var entries []PortEntry

	filename := "/proc/net/" + proto
	f, err := os.Open(filename)
	if err != nil {
		return entries
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// Skip header line
	if !sc.Scan() {
		return entries
	}

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// /proc/net/tcp format:
		// sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
		// 0: 00000000:0050 00000000:0000 01 00000000:00000000 00:00000000 00000000     0 0 12345
		// local_address is little-endian u32 + port

		localHex := fields[1]
		remoteHex := fields[2]
		stateCode, _ := strconv.ParseUint(fields[3], 16, 32)
		uid, _ := strconv.ParseUint(fields[4], 10, 64)
		_ = uid // reserved for future use
		inodeField := fields[9]

		localAddr, localPort := decodeHexAddr(localHex)
		remoteAddr, remotePort := decodeHexAddr(remoteHex)

		state := tcpStateName(uint32(stateCode))

		// Find PID via inode
		pid := InodeToPID(inodeField)
		procName := ""
		if pid > 0 {
			procName = PIDToName(pid)
		}

		entries = append(entries, PortEntry{
			LocalAddr:   localAddr,
			LocalPort:   localPort,
			RemoteAddr:  remoteAddr,
			RemotePort:  remotePort,
			State:       state,
			StateCode:   uint32(stateCode),
			PID:         pid,
			ProcessName: procName,
		})
	}

	return entries
}

func decodeHexAddr(hexStr string) (string, uint32) {
	parts := strings.Split(hexStr, ":")
	if len(parts) != 2 {
		return "", 0
	}

	// IP part is little-endian hex
	ipHex := parts[0]
	portHex := parts[1]

	port, _ := strconv.ParseUint(portHex, 16, 32)

	// Convert little-endian hex to IP
	ip, _ := hex.DecodeString(ipHex)
	if len(ip) == 4 {
		// Little-endian: reverse bytes
		addr := net.IP{ip[3], ip[2], ip[1], ip[0]}
		return addr.String(), uint32(port)
	} else if len(ip) == 16 {
		// IPv6
		return net.IP(ip).String(), uint32(port)
	}

	return "", uint32(port)
}

func tcpStateName(code uint32) string {
	switch code {
	case 1:
		return "ESTABLISHED"
	case 2:
		return "SYN_SENT"
	case 3:
		return "SYN_RECV"
	case 4:
		return "FIN_WAIT1"
	case 5:
		return "FIN_WAIT2"
	case 6:
		return "TIME_WAIT"
	case 7:
		return "CLOSE"
	case 8:
		return "CLOSE_WAIT"
	case 9:
		return "LAST_ACK"
	case 10:
		return "LISTEN"
	case 11:
		return "CLOSING"
	default:
		return "UNKNOWN"
	}
}

// InodeToPID finds the PID that owns a socket with the given inode number.
func InodeToPID(inodeStr string) uint32 {
	inode, err := strconv.ParseUint(inodeStr, 10, 64)
	if err != nil {
		return 0
	}

	// Scan /proc/*/fd/* looking for socket:[<inode>]
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pidStr := e.Name()
		if pidStr[0] < '0' || pidStr[0] > '9' {
			continue
		}
		pid, err := strconv.ParseUint(pidStr, 10, 64)
		if err != nil {
			continue
		}

		fdDir := "/proc/" + pidStr + "/fd"
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(fdDir + "/" + fd.Name())
			if err != nil {
				continue
			}
			// Socket link format: "socket:[12345]"
			if strings.HasPrefix(link, "socket:[") {
				start := len("socket:[")
				end := len(link) - 1
				if end > start {
					socketInode, err := strconv.ParseUint(link[start:end], 10, 64)
					if err == nil && socketInode == inode {
						return uint32(pid)
					}
				}
			}
		}
	}
	return 0
}

// PIDToName returns the process name for a given PID.
func PIDToName(pid uint32) string {
	commPath := "/proc/" + strconv.FormatUint(uint64(pid), 10) + "/comm"
	data := ReadFile(commPath)
	if data != "" {
		return strings.TrimSpace(data)
	}
	// Fallback: read cmdline
	cmdPath := "/proc/" + strconv.FormatUint(uint64(pid), 10) + "/cmdline"
	data = ReadFile(cmdPath)
	if data != "" {
		parts := strings.Split(data, "\x00")
		if len(parts) > 0 && parts[0] != "" {
			// Extract basename
			parts2 := strings.Split(parts[0], "/")
			return parts2[len(parts2)-1]
		}
	}
	return ""
}
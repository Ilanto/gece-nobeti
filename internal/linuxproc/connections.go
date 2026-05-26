package linuxproc

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConnectionEntry represents one TCP/UDP connection.
type ConnectionEntry struct {
	Protocol   string
	LocalAddr  string
	LocalPort  uint16
	RemoteAddr string
	RemotePort uint16
	State      string
	PID        uint32
	Process    string
}

var tcpStates = map[string]string{
	"01": "ESTABLISHED", "02": "SYN_SENT", "03": "SYN_RECV",
	"04": "FIN_WAIT1", "05": "FIN_WAIT2", "06": "TIME_WAIT",
	"07": "CLOSE", "08": "CLOSE_WAIT", "09": "LAST_ACK",
	"0A": "LISTEN", "0B": "CLOSING",
}

func hexToIPv4(h string) string {
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != 4 {
		return "0.0.0.0"
	}
	ip := binary.LittleEndian.Uint32(b)
	return fmt.Sprintf("%d.%d.%d.%d", ip&0xff, (ip>>8)&0xff, (ip>>16)&0xff, (ip>>24)&0xff)
}

func hexToPort(h string) uint16 {
	v, _ := strconv.ParseUint(h, 16, 16)
	return uint16(v)
}

func buildInodeMap() map[string][2]string {
	m := make(map[string][2]string)
	fdPaths, _ := filepath.Glob("/proc/[0-9]*/fd/*")
	for _, fdPath := range fdPaths {
		link, err := os.Readlink(fdPath)
		if err != nil || !strings.HasPrefix(link, "socket:[") {
			continue
		}
		inode := link[8 : len(link)-1]
		parts := strings.Split(fdPath, "/")
		if len(parts) < 3 {
			continue
		}
		pid := parts[2]
		name := ""
		if b, err := os.ReadFile("/proc/" + pid + "/comm"); err == nil {
			name = strings.TrimSpace(string(b))
		}
		m[inode] = [2]string{pid, name}
	}
	return m
}

func parseNetFile(path, proto string, inodeMap map[string][2]string) []ConnectionEntry {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var entries []ConnectionEntry
	sc := bufio.NewScanner(f)
	sc.Scan() // skip header

	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 10 {
			continue
		}

		localParts := strings.Split(fields[1], ":")
		remoteParts := strings.Split(fields[2], ":")
		if len(localParts) != 2 || len(remoteParts) != 2 {
			continue
		}

		stateHex := strings.ToUpper(fields[3])
		state := tcpStates[stateHex]
		if state == "" {
			state = stateHex
		}

		inode := fields[9]
		pid := uint32(0)
		procName := ""
		if info, ok := inodeMap[inode]; ok {
			if v, err := strconv.ParseUint(info[0], 10, 32); err == nil {
				pid = uint32(v)
			}
			procName = info[1]
		}

		var localIP, remoteIP string
		if len(localParts[0]) == 32 {
			// IPv6
			b, _ := hex.DecodeString(localParts[0])
			localIP = net.IP(b).String()
			b2, _ := hex.DecodeString(remoteParts[0])
			remoteIP = net.IP(b2).String()
		} else {
			localIP = hexToIPv4(localParts[0])
			remoteIP = hexToIPv4(remoteParts[0])
		}

		entries = append(entries, ConnectionEntry{
			Protocol:   proto,
			LocalAddr:  localIP,
			LocalPort:  hexToPort(localParts[1]),
			RemoteAddr: remoteIP,
			RemotePort: hexToPort(remoteParts[1]),
			State:      state,
			PID:        pid,
			Process:    procName,
		})
	}
	return entries
}

// CollectConnections reads all TCP/UDP connections from /proc/net.
func CollectConnections() []ConnectionEntry {
	inodeMap := buildInodeMap()
	var all []ConnectionEntry
	all = append(all, parseNetFile("/proc/net/tcp",  "tcp",  inodeMap)...)
	all = append(all, parseNetFile("/proc/net/tcp6", "tcp6", inodeMap)...)
	all = append(all, parseNetFile("/proc/net/udp",  "udp",  inodeMap)...)
	return all
}

//go:build linux

package controller

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	ErrProtectedProcess = errors.New("cannot kill/protect/suspend protected process")
	ErrProcessNotFound  = errors.New("process not found")
)

// SafetyManager protects critical system processes from being killed/suspended.
type SafetyManager struct {
	protected    map[string]bool // lowercase process names that are protected
	protectedPIDs map[uint32]bool // specific PIDs that are protected (e.g., PID 1)
}

// NewSafetyManager creates a new safety manager with the protected process list from config.
func NewSafetyManager(protectedProcesses []string) *SafetyManager {
	m := make(map[string]bool)
	pids := make(map[uint32]bool)

	for _, p := range protectedProcesses {
		// Check if it's a PID (numeric string)
		if pid, err := strconv.ParseUint(p, 10, 32); err == nil {
			pids[uint32(pid)] = true
		} else {
			m[normalize(p)] = true
		}
	}

	return &SafetyManager{
		protected:    m,
		protectedPIDs: pids,
	}
}

// IsProtected returns true if the given PID is a protected system process.
func (s *SafetyManager) IsProtected(pid uint32) bool {
	// PID 1 (init/systemd) is always protected
	if pid == 1 {
		return true
	}
	if s.protectedPIDs[pid] {
		return true
	}
	return false
}

// ValidateAction checks if an action is allowed on a given PID.
// Returns an error if the process is protected.
func (s *SafetyManager) ValidateAction(pid uint32, action string) error {
	if s.IsProtected(pid) {
		return ErrProtectedProcess
	}
	return nil
}

// Controller wraps OS-level process control.
type Controller struct {
	protected map[string]bool // lowercase process names
	safety    *SafetyManager
}

// NewController returns a new Controller.
func NewController(protected []string) *Controller {
	m := make(map[string]bool)
	for _, p := range protected {
		m[normalize(p)] = true
	}
	return &Controller{protected: m, safety: NewSafetyManager(protected)}
}

// NewControllerWithSafety creates a new Controller with an explicit SafetyManager.
func NewControllerWithSafety(protected []string, safety *SafetyManager) *Controller {
	m := make(map[string]bool)
	for _, p := range protected {
		m[normalize(p)] = true
	}
	return &Controller{protected: m, safety: safety}
}

// IsProtected checks if a process name/PID is protected.
func (c *Controller) IsProtected(pid uint32) bool {
	// PID 1 (init/systemd) is always protected
	if pid == 1 {
		return true
	}
	name := procName(pid)
	if c.protected[normalize(name)] {
		return true
	}
	if c.safety != nil && c.safety.IsProtected(pid) {
		return true
	}
	return false
}

// ValidateAction checks if an action is allowed on a given PID.
// Returns an error if the process is protected.
func (c *Controller) ValidateAction(pid uint32, action string) error {
	if c.IsProtected(pid) {
		return ErrProtectedProcess
	}
	return nil
}

func normalize(s string) string {
	// lowercase, trim
	return strings.ToLower(strings.TrimSpace(s))
}

// Kill sends SIGKILL to the process.
func (c *Controller) Kill(pid uint32) error {
	if c.IsProtected(pid) {
		return ErrProtectedProcess
	}
	proc, _ := os.FindProcess(int(pid))
	err := proc.Kill()
	if err != nil {
		return err
	}
	return nil
}

// Suspend sends SIGSTOP to the process.
func (c *Controller) Suspend(pid uint32) error {
	if c.IsProtected(pid) {
		return ErrProtectedProcess
	}
	err := syscall.Kill(int(pid), syscall.SIGSTOP)
	return err
}

// Resume sends SIGCONT to the process.
func (c *Controller) Resume(pid uint32) error {
	return syscall.Kill(int(pid), syscall.SIGCONT)
}

// SetPriority sets the nice value (-20 to 19, lower = higher priority).
func (c *Controller) SetPriority(pid uint32, nice int) error {
	if nice < -20 {
		nice = -20
	}
	if nice > 19 {
		nice = 19
	}
	return syscall.Setpriority(syscall.PRIO_PROCESS, int(pid), nice)
}

// SetAffinity sets CPU affinity mask using sched_setaffinity.
func (c *Controller) SetAffinity(pid uint32, mask []uint64) error {
	// Use unix package for CPU affinity since syscall doesn't provide CPUSet
	// But since we can only use syscall, implement via direct syscall
	return setAffinitySyscall(pid, mask)
}

// setAffinitySyscall calls sched_setaffinity via syscall.
func setAffinitySyscall(pid uint32, mask []uint64) error {
	// Determine size based on mask length
	size := len(mask) * 8 // bytes
	
	// Use unix.RawSyscall to call sched_setaffinity
	// sched_setaffinity(pid_t pid, size_t cpusetsize, const cpu_set_t *mask)
	// On Linux, cpu_set_t is a bitmask where each bit represents a CPU
	
	// Get the first cpu_set_t pointer
	var cpuMask [16]uint64 // enough for up to 1024 CPUs
	for i := 0; i < len(mask) && i < len(cpuMask); i++ {
		cpuMask[i] = mask[i]
	}
	
	_, _, errno := syscall.Syscall(
		syscall.SYS_SCHED_SETAFFINITY,
		uintptr(pid),
		uintptr(size),
		uintptr(unsafe.Pointer(&cpuMask[0])),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// procName reads /proc/<pid>/comm safely.
func procName(pid uint32) string {
	data, err := os.ReadFile("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/comm")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
package guard

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

// The darwin sampler calls __proc_info (syscall 336) through the standard
// library — no cgo, no new module (decision darwin-proc-info-no-cgo):
// PROC_INFO_CALL_LISTPIDS for the process list, PROC_PIDTBSDINFO for each
// pid's status, ppid and pgid, and PROC_INFO_CALL_PIDRUSAGE with
// RUSAGE_INFO_V2 for ri_phys_footprint.
const (
	rcSysProcInfo      = 336
	rcCallListPids     = 1
	rcCallPidInfo      = 2
	rcCallPidRusage    = 9
	rcAllPids          = 1
	rcPidTBSDInfo      = 3
	rcBSDInfoSize      = 136 // sizeof(struct proc_bsdinfo)
	rcBSDStatusOff     = 4   // pbi_status
	rcBSDPpidOff       = 16  // pbi_ppid
	rcBSDPgidOff       = 100 // pbi_pgid
	rcStatusZombie     = 5   // SZOMB
	rcRusageInfoV2     = 2
	rcRusageV2Size     = 160 // sizeof(struct rusage_info_v2)
	rcPhysFootprintOff = 72  // ri_phys_footprint
)

func rcProcInfo(call, pid, flavor uintptr, buf []byte) (int, bool) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&buf[0])
	}
	r, _, errno := syscall.Syscall6(rcSysProcInfo, call, pid, flavor, 0, uintptr(p), uintptr(len(buf)))
	return int(r), errno == 0
}

// rcProcs lists every process with its ppid, pgid and whether it is a
// zombie. A process that exits between the list and its own read drops out.
func rcProcs() []rcProc {
	buf := make([]byte, 4*4096)
	for {
		n, ok := rcProcInfo(rcCallListPids, rcAllPids, 0, buf)
		if !ok {
			return nil
		}
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	info := make([]byte, rcBSDInfoSize)
	var procs []rcProc
	for i := 0; i+4 <= len(buf); i += 4 {
		pid := int(int32(binary.LittleEndian.Uint32(buf[i:])))
		if pid <= 0 {
			continue
		}
		if n, ok := rcProcInfo(rcCallPidInfo, uintptr(pid), rcPidTBSDInfo, info); !ok || n != rcBSDInfoSize {
			continue
		}
		procs = append(procs, rcProc{
			pid:    pid,
			ppid:   int(binary.LittleEndian.Uint32(info[rcBSDPpidOff:])),
			pgid:   int(binary.LittleEndian.Uint32(info[rcBSDPgidOff:])),
			zombie: binary.LittleEndian.Uint32(info[rcBSDStatusOff:]) == rcStatusZombie,
		})
	}
	return procs
}

// rcFootprint is pid's ri_phys_footprint in bytes: its resident, compressed
// and swapped dirty memory; 0 when it cannot be read.
func rcFootprint(pid int) uint64 {
	buf := make([]byte, rcRusageV2Size)
	if _, ok := rcProcInfo(rcCallPidRusage, uintptr(pid), rcRusageInfoV2, buf); !ok {
		return 0
	}
	return binary.LittleEndian.Uint64(buf[rcPhysFootprintOff:])
}

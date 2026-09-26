package guard

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

// Offsets into <sys/sysctl.h>'s struct kinfo_proc, measured with offsetof
// on this platform's SDK -- identical on arm64 and x86_64.
const (
	rrKinfoSize = 648 // sizeof(struct kinfo_proc)
	rrKinfoPid  = 40  // kp_proc.p_pid
	rrKinfoPpid = 560 // kp_eproc.e_ppid
	rrKinfoPgid = 564 // kp_eproc.e_pgid
)

// rrProcTable is one read of the process table; a failed read is an error,
// never an empty table. The bash read it as `pgrep ... || true`, and an
// empty table there left a live survivor unnamed -- a verdict where the
// contract requires "unverifiable" (survivor-detection-deterministic; panel
// round 0, F7/F15) -- so run-reproducer answers a failed read with its
// cannot-answer exit 4 instead. It is read in-process through sysctl
// kern.proc.all -- the source ps(1) reads -- never by exec'ing ps: under
// load one ps exec measured 360-900ms, longer than a detaching fixture's
// parent lives, so the snapshots meant to be rrPoll apart fell on either
// side of it and a child that calls setsid went unnamed (case 13).
//
// The buffer holds kern.maxproc records, the most the table can hold, so the
// read cannot fail for want of room. syscall.Sysctl sizes its buffer by a
// probe instead, and under fork load the table outgrew that probe before the
// read: ENOMEM, an empty table, and a lost group snapshot (cases 16, 18).
func rrProcTable() ([]rrProc, error) {
	maxproc, err := syscall.SysctlUint32("kern.maxproc")
	if err != nil {
		return nil, fmt.Errorf("sysctl kern.maxproc: %w", err)
	}
	mib := [3]int32{1, 14, 0} // CTL_KERN, KERN_PROC, KERN_PROC_ALL
	b := make([]byte, int(maxproc)*rrKinfoSize)
	n := uintptr(len(b))
	if _, _, e := syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), uintptr(len(mib)),
		uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&n)), 0, 0); e != 0 {
		return nil, fmt.Errorf("sysctl kern.proc.all: %w", e)
	}
	procs := make([]rrProc, 0, int(n)/rrKinfoSize)
	for off := 0; off+rrKinfoSize <= int(n); off += rrKinfoSize {
		field := func(at int) int { return int(int32(binary.NativeEndian.Uint32(b[off+at:]))) }
		procs = append(procs, rrProc{pid: field(rrKinfoPid), ppid: field(rrKinfoPpid), pgid: field(rrKinfoPgid)})
	}
	return procs, nil
}

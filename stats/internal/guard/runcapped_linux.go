package guard

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// rcProcs lists every process under /proc with its ppid, pgid and whether
// it is a zombie, read from /proc/<pid>/stat: the fields after the
// parenthesised command name are state, ppid and pgrp. A process that exits
// between the listing and its own read drops out.
func rcProcs() []rcProc {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	var procs []rcProc
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		b, err := os.ReadFile("/proc/" + e.Name() + "/stat")
		if err != nil {
			continue
		}
		s := string(b)
		fields := strings.Fields(s[strings.LastIndexByte(s, ')')+1:])
		if len(fields) < 3 {
			continue
		}
		ppid, err1 := strconv.Atoi(fields[1])
		pgid, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil {
			continue
		}
		procs = append(procs, rcProc{pid: pid, ppid: ppid, pgid: pgid, zombie: fields[0] == "Z"})
	}
	return procs
}

// rcFootprint is pid's resident plus swapped memory in bytes — VmRSS +
// VmSwap of /proc/<pid>/status; 0 when it cannot be read.
func rcFootprint(pid int) uint64 {
	f, err := os.Open("/proc/" + strconv.Itoa(pid) + "/status")
	if err != nil {
		return 0
	}
	defer f.Close()
	var kb uint64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), ":")
		if !ok || (k != "VmRSS" && k != "VmSwap") {
			continue
		}
		if fs := strings.Fields(v); len(fs) > 0 {
			n, _ := strconv.ParseUint(fs[0], 10, 64)
			kb += n
		}
	}
	return kb << 10
}

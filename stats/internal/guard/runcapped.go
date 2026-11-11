package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// runCapped is `flow-guard run-capped` — run a command and kill its whole
// process tree once the tree's summed physical footprint exceeds a memory
// cap, so a suite, build or load reproduction a flow run starts cannot
// overload the machine.
//
// Usage: flow-guard run-capped [-m <MB>] -- <cmd...>
//
// Without -m the cap is the body of `<cwd>/.flow/project.md`'s
// `## resource cap` (project-get): a plain integer of MB, nothing else.
// Absent means 4096 MB.
//
// The tree is every process whose ppid chain reaches the command, plus every
// member of the command's process group — so a child that calls setsid (a
// Playwright browser) is measured and killed while its parent lives. A pid
// found in the tree once stays in the kill set until the process list no
// longer lists it or lists it as a zombie, so such a session leader whose
// parent died is still killed; a remembered pid is signalled only while the
// list still holds it (kan-1103 self-review, F1). A
// process that reparents to PID 1 (a Gradle daemon) escapes it: the wrapper
// therefore sets GRADLE_OPTS `-Dorg.gradle.daemon=false`, so Gradle builds in
// its client's tree.
//
// The command's stdin is /dev/null: the command runs as a background group
// and would stop on a tty read. The cap is per command; a run-capped
// started under another one (RUN_CAPPED_PGID set) execs its command in place,
// so the outer cap covers it.
//
// Exit codes: the command's own status; 137 when the cap killed the tree
// (stderr carries `CAPPED: <cap>MB (tree footprint <n>MB)`); 130 on SIGINT, 143 on
// SIGTERM, 129 on SIGHUP (tree killed); 2 usage, or a `## resource cap` that
// project-get refuses or that is not a plain MB integer.
//
// ponytail: polls the footprint once a second, so a spike inside one second can
// overshoot the cap; cgroups/ulimit are not available on darwin.
//
// THE FOOTPRINT, NEVER RSS (decision cap-physical-footprint). macOS
// compresses an idle process's pages out of its resident set — a 128 MB
// allocator left idle read 142 MB of RSS, then 8 MB ten seconds later, while
// its physical footprint held at 135 MB — so a tree summed by RSS escapes a
// cap it is far over. Each member is measured by its physical footprint:
// compressed and swapped pages included (runcapped_darwin.go,
// runcapped_linux.go), and the process list is read without forking `ps`.
func init() { Registry["run-capped"] = runCapped }

const rcUsage = "usage: flow-guard run-capped [-m <MB>] -- <cmd...>"

// rcProc is one process-table row the tree is built from.
type rcProc struct {
	pid, ppid, pgid int
	zombie          bool
}

func runCapped(args []string, env Env, stdout, stderr io.Writer) int {
	capMB := ""
	if len(args) >= 2 && args[0] == "-m" {
		capMB, args = args[1], args[2:]
	}
	if len(args) == 0 || args[0] != "--" || len(args) == 1 {
		fmt.Fprintln(stderr, rcUsage)
		return 2
	}
	args = args[1:]
	environ := icChildEnv(env)
	path, ok := lookPath(env, args[0])
	if strings.Contains(args[0], "/") || !ok {
		path = args[0]
	}
	if env.Getenv("RUN_CAPPED_PGID") != "" {
		err := syscall.Exec(path, args, environ)
		fmt.Fprintf(stderr, "run-capped: %s: %v\n", args[0], err)
		return 127
	}
	if capMB == "" {
		var body, errb bytes.Buffer
		switch Registry["project-get"]([]string{env.Dir, "resource cap"}, env, &body, &errb) {
		case 0:
			capMB = strings.Join(strings.Fields(body.String()), " ")
		case 1:
			capMB = "4096"
		default:
			fmt.Fprintf(stderr, "run-capped: project-get failed: %s\n", strings.TrimRight(errb.String(), "\n"))
			return 2
		}
	}
	limit, err := strconv.Atoi(capMB)
	if err != nil || strings.TrimLeft(capMB, "0123456789") != "" {
		fmt.Fprintf(stderr, "run-capped: cap must be a plain integer of MB: %s\n", capMB)
		return 2
	}

	gradle := ""
	for _, kv := range environ {
		if v, ok := strings.CutPrefix(kv, "GRADLE_OPTS="); ok {
			gradle = v
		}
	}
	cmd := exec.Command(path, args[1:]...)
	cmd.Args[0] = args[0]
	cmd.Dir = env.Dir
	cmd.Env = append(append([]string{}, environ...), "GRADLE_OPTS="+gradle+" -Dorg.gradle.daemon=false", "RUN_CAPPED_PGID="+strconv.Itoa(os.Getpid()))
	cmd.Stdout, cmd.Stderr = stdout, stderr
	// The command leads its own process group, as the bash's `set -m` job did.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sigs)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(stderr, "run-capped: %s: %v\n", args[0], err)
		return 127
	}
	pg := cmd.Process.Pid
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		if b := rcTreeFootprint(pg); b > uint64(limit)<<20 {
			rcKillTree(pg)
			<-done
			fmt.Fprintf(stderr, "CAPPED: %dMB (tree footprint %dMB)\n", limit, b>>20)
			return 137
		}
		select {
		case err := <-done:
			return exitCode(err)
		case sig := <-sigs:
			rcKillTree(pg)
			<-done
			return map[os.Signal]int{syscall.SIGINT: 130, syscall.SIGTERM: 143, syscall.SIGHUP: 129}[sig]
		case <-tick.C:
		}
	}
}

// rcTree is the live (non-zombie) pids of pg's tree in the current process
// table.
func rcTree(pg int) []int { return rcTreeIn(pg, rcProcs()) }

// rcTreeIn is the live (non-zombie) pids of pg's tree in procs: descendants
// of pg by ppid, plus members of group pg.
func rcTreeIn(pg int, procs []rcProc) []int {
	in := map[int]bool{pg: true}
	for _, p := range procs {
		if p.pgid == pg {
			in[p.pid] = true
		}
	}
	for more := true; more; {
		more = false
		for _, p := range procs {
			if !in[p.pid] && in[p.ppid] {
				in[p.pid], more = true, true
			}
		}
	}
	var tree []int
	for _, p := range procs {
		if in[p.pid] && !p.zombie {
			tree = append(tree, p.pid)
		}
	}
	return tree
}

// rcTreeFootprint is the summed physical footprint, in bytes, of pg's tree.
func rcTreeFootprint(pg int) uint64 {
	var sum uint64
	for _, pid := range rcTree(pg) {
		sum += rcFootprint(pid)
	}
	return sum
}

// rcKillTree is the bash's kill_tree: TERM to the group and every member,
// up to 2 s for them to go, then KILL, re-signalled each pass so a child
// forked after a snapshot is killed too. Every pass acts on every pid any
// pass found in the tree, not on the tree re-derived: a session leader whose
// parent died of the TERM is reparented out of the tree and would otherwise
// never get the KILL. A remembered pid is signalled only while rcProcs still
// lists it alive, so a pid gone or a zombie is never signalled (kan-1103
// self-review, F1).
// ponytail: re-listing does not detect pid reuse — a remembered pid that
// exits and is reassigned within the kill window (seconds; pids allocate
// incrementally) is signalled; match start time or lineage if that bites.
func rcKillTree(pg int) {
	seen := map[int]bool{}
	live := func() []int {
		procs := rcProcs()
		for _, pid := range rcTreeIn(pg, procs) {
			seen[pid] = true
		}
		var pids []int
		for _, p := range procs {
			if seen[p.pid] && !p.zombie {
				pids = append(pids, p.pid)
			}
		}
		return pids
	}
	signalAll := func(sig syscall.Signal, pids []int) {
		_ = syscall.Kill(-pg, sig)
		for _, pid := range pids {
			if pid > 0 {
				_ = syscall.Kill(pid, sig)
			}
		}
	}
	pids := live()
	if len(pids) == 0 {
		return
	}
	signalAll(syscall.SIGTERM, pids)
	for i := 0; i < 4; i++ {
		if len(live()) == 0 {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	for pids := live(); len(pids) > 0; pids = live() {
		signalAll(syscall.SIGKILL, pids)
		time.Sleep(100 * time.Millisecond)
	}
}

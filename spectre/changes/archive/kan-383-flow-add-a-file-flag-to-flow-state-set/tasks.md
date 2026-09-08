# kan-383-flow-add-a-file-flag-to-flow-state-set

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

- [x] 1. Add `-file <path>` to `flow state set` — CLI, usage text, tests

  - [x] **Step 1: write the failing tests in `stats/cmd/flow/state_test.go`**

  Five new top-level test functions, no `t.Run` subtests, so the package's
  `=== RUN   Test` count moves by exactly their number. They use the file's
  existing helpers (`gitRepo`, `isolatedStateRoot`, `deadPortAddr`, `run`)
  unchanged, and follow `TestStateSetRejectsNullBody`'s shape: drive `run`
  directly, assert the exit code, assert the fallback file was or was not
  written. Every error-path test also asserts stderr names the offending
  file path — that assertion is what makes each test red before Step 3,
  since a not-yet-existing flag also exits 2.

  ```go verified:helper names and call shapes read against stats/cmd/flow/state_test.go on this branch (gitRepo, isolatedStateRoot, deadPortAddr, run, fallback.ProjectKey, fallback.ReadStateFile)
  func TestStateSetFileFlagReadsRecordFromFile(t *testing.T) {
  	repo := gitRepo(t)
  	isolatedStateRoot(t)
  	path := filepath.Join(t.TempDir(), "record.json")
  	if err := os.WriteFile(path, []byte(`{"state":"IN_PROGRESS"}`), 0o644); err != nil {
  		t.Fatal(err)
  	}

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(),
  		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "-file", path, "kan-383"},
  		strings.NewReader(""), // stdin must be ignored when -file is given
  		&stdout, &stderr)

  	if code != 0 {
  		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
  	}
  	projectKey, _, err := fallback.ProjectKey(repo)
  	if err != nil {
  		t.Fatalf("ProjectKey: %v", err)
  	}
  	body, err := fallback.ReadStateFile(fallback.StateFilePath(projectKey, "kan-383"))
  	if err != nil {
  		t.Fatalf("ReadStateFile: %v", err)
  	}
  	var rec map[string]json.RawMessage
  	if err := json.Unmarshal(body, &rec); err != nil {
  		t.Fatalf("fallback record is not JSON: %v", err)
  	}
  	if string(rec["state"]) != `"IN_PROGRESS"` {
  		t.Errorf("state = %s, want the file's own value", rec["state"])
  	}
  }

  func TestStateSetFileFlagMissingFileIsUsageError(t *testing.T) {
  	repo := gitRepo(t)
  	isolatedStateRoot(t)
  	path := filepath.Join(t.TempDir(), "absent.json")

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(),
  		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "-file", path, "kan-383"},
  		strings.NewReader(`{"state":"IN_PROGRESS"}`),
  		&stdout, &stderr)

  	if code != 2 {
  		t.Fatalf("exit code = %d, want 2 (an unreadable -file is a usage error, not a store failure); stderr:\n%s", code, stderr.String())
  	}
  	if !strings.Contains(stderr.String(), path) {
  		t.Errorf("stderr must name the unreadable path, got:\n%s", stderr.String())
  	}
  	projectKey, _, err := fallback.ProjectKey(repo)
  	if err != nil {
  		t.Fatalf("ProjectKey: %v", err)
  	}
  	if _, err := fallback.ReadStateFile(fallback.StateFilePath(projectKey, "kan-383")); err == nil {
  		t.Error("an unreadable -file must never fall back — no state file may exist")
  	}
  }

  func TestStateSetFileFlagNonObjectFileIsUsageError(t *testing.T) {
  	repo := gitRepo(t)
  	isolatedStateRoot(t)
  	path := filepath.Join(t.TempDir(), "record.json")
  	if err := os.WriteFile(path, []byte(`[1,2,3]`), 0o644); err != nil {
  		t.Fatal(err)
  	}

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(),
  		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "-file", path, "kan-383"},
  		strings.NewReader(""),
  		&stdout, &stderr)

  	if code != 2 {
  		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
  	}
  	if !strings.Contains(stderr.String(), path) {
  		t.Errorf("stderr must name the offending file, got:\n%s", stderr.String())
  	}
  }

  func TestStateSetFileFlagOversizeFileIsUsageError(t *testing.T) {
  	repo := gitRepo(t)
  	isolatedStateRoot(t)
  	path := filepath.Join(t.TempDir(), "record.json")
  	big := `{"pad":"` + strings.Repeat("a", maxStdinBytes) + `"}`
  	if err := os.WriteFile(path, []byte(big), 0o644); err != nil {
  		t.Fatal(err)
  	}

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(),
  		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "-file", path, "kan-383"},
  		strings.NewReader(""),
  		&stdout, &stderr)

  	if code != 2 {
  		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
  	}
  	if !strings.Contains(stderr.String(), path) {
  		t.Errorf("stderr must name the oversize file, got:\n%s", stderr.String())
  	}
  }

  func TestStateGetFileFlagIsUnknownFlag(t *testing.T) {
  	repo := gitRepo(t)

  	var stdout, stderr bytes.Buffer
  	code := run(context.Background(),
  		[]string{"state", "get", "-file", "/tmp/unused.json", "-C", repo, "kan-383"},
  		strings.NewReader(""), &stdout, &stderr)

  	if code != 2 {
  		t.Fatalf("exit code = %d, want 2 (-file is state set only); stderr:\n%s", code, stderr.String())
  	}
  }
  ```

  - [x] **Step 2: run them to verify they fail**

  ```bash unverified:the named tests resolve only once Step 1 has written them
  cd stats && go test ./cmd/flow -run 'TestStateSetFileFlag|TestStateGetFileFlag' -count=1 -v
  ```

  Expected: the four `-file` tests and the `state get` test FAIL — the flag
  does not exist yet, so the success test exits 2 instead of 0, and the
  error-path tests pass their exit-code check but fail the new
  stderr-names-the-path assertion.

  - [x] **Step 3: implement the flag in `runStateSet` (`stats/cmd/flow/state.go`)**

  Register `-file` on `runStateSet`'s own `flag.FlagSet` before
  `parseStateFlags` adds `-addr`/`-timeout`/`-C` and parses that same set —
  `parseStateFlags` takes the set as a parameter precisely so a caller can
  pre-register, and a flag registered only here never exists for
  `state get`. Read the body from the file when the flag carries a value,
  from stdin otherwise, then let the existing cap and `isJSONObject`
  checks stand unchanged, with the two error messages branching on the
  source so every existing message stays byte-identical. Everything from
  `stampUpdatedAt` onward is untouched.

  ```go unverified:not yet compiled — Step 5's go vet and tests confirm it
  func runStateSet(ctx context.Context, args []string, stdin io.Reader, _, stderr io.Writer) int {
  	fset := flag.NewFlagSet("flow state set", flag.ContinueOnError)
  	// -file registers here, before parseStateFlags adds -addr/-timeout/-C
  	// and parses this same FlagSet, so the flag exists on state set only:
  	// state get's FlagSet never sees it, and `state get -file` stays a
  	// usage error.
  	file := fset.String("file", "", "read the record from this file instead of stdin")
  	f, err := parseStateFlags(fset, args, stderr)
  	if err != nil {
  		if errors.Is(err, flag.ErrHelp) {
  			return 0
  		}
  		fmt.Fprintf(stderr, "flow: %v\n", err)
  		fmt.Fprint(stderr, stateUsage)
  		return 2
  	}

  	var body []byte
  	if *file != "" {
  		// A file the caller named but that cannot be read is a local input
  		// error in state-file.md's sense: reported, exit 2, never the
  		// fallback path, never the network.
  		body, err = os.ReadFile(*file)
  		if err != nil {
  			fmt.Fprintf(stderr, "flow: read state from %s: %v\n", *file, err)
  			return 2
  		}
  	} else {
  		body, err = io.ReadAll(io.LimitReader(stdin, maxStdinBytes+1))
  		if err != nil {
  			fmt.Fprintf(stderr, "flow: read state from stdin: %v\n", err)
  			return 1
  		}
  	}
  	if len(body) > maxStdinBytes {
  		if *file != "" {
  			fmt.Fprintf(stderr, "flow: state in %s exceeds %d bytes\n", *file, maxStdinBytes)
  		} else {
  			fmt.Fprintf(stderr, "flow: state on stdin exceeds %d bytes\n", maxStdinBytes)
  		}
  		return 2
  	}
  	if !isJSONObject(body) {
  		if *file != "" {
  			fmt.Fprintf(stderr, "flow: state in %s must be a JSON object\n", *file)
  		} else {
  			fmt.Fprintln(stderr, "flow: state on stdin must be a JSON object")
  		}
  		return 2
  	}

  	// ... unchanged from here: stampUpdatedAt, validateWorktreeMergeBases,
  	// fallback.ProjectKey, putChange and the fallback branch stay exactly
  	// as they are today.
  ```

  - [x] **Step 4: update the usage strings**

  `stateUsage` in `stats/cmd/flow/state.go` — the `state set` line gains the
  flag and the stdin sentence gains the file source:

  ```text verified:current strings read against stats/cmd/flow/state.go's stateUsage and stats/cmd/flow/main.go's usage constant on this branch
  usage: flow state get     [-addr url] [-timeout dur] [-C dir] <name>
         flow state set     [-addr url] [-timeout dur] [-C dir] [-file path] <name>

  state set reads the change's whole state as JSON from the file named by
  -file when it is given, and from stdin otherwise.
  ```

  `usage` in `stats/cmd/flow/main.go` — the `state set` line becomes:

  ```text verified:current line read against stats/cmd/flow/main.go's usage constant on this branch
    state set <name>    write the change's whole state, from -file or stdin
  ```

  - [x] **Step 5: run this task's verification**

  ```bash verified:commands as named in .flow/project.md's ## lint list (the Go entries) plus the targeted-test selector
  cd stats && gofmt -l .
  cd stats && go vet ./...
  cd stats && go test ./cmd/flow -run 'TestStateSetFileFlag|TestStateGetFileFlag' -count=1 -v
  cd stats && go test ./cmd/flow -count=1 -v | grep -c '^=== RUN   Test'
  ```

  Expected: `gofmt -l` prints nothing, `go vet` is clean, the new tests
  PASS, and the last command prints the `after` count declared in this
  task's `**Baseline:**` field.

  - [x] **Step 6: commit**

  ```bash verified:subject shape matches the commit this repository's archived kan-224 plan declared for its CLI task, and check-task-commit-fields.sh's field grammar
  git add stats/cmd/flow/state.go stats/cmd/flow/main.go stats/cmd/flow/state_test.go
  git commit -m "feat(flow-cli): read state set's record from a -file path"
  ```

**Files:** `stats/cmd/flow/state.go`, `stats/cmd/flow/main.go`, `stats/cmd/flow/state_test.go`
**Tests:** `TestStateSetFileFlagReadsRecordFromFile`, `TestStateSetFileFlagMissingFileIsUsageError`, `TestStateSetFileFlagNonObjectFileIsUsageError`, `TestStateSetFileFlagOversizeFileIsUsageError`, `TestStateGetFileFlagIsUnknownFlag`
**Regression:** reverting this task's commit removes the `-file` flag, so each of the five tests above fails — the success test back to exit 2, the error-path tests back to an unnamed-path stderr, and `state get -file` back to accepting an undefined flag differently; the stdin path is untouched, so every pre-existing test still passes
**Baseline:** before=200 after=205
<!-- measured: cd stats && go test ./cmd/flow -count=1 -v | grep -c '^=== RUN   Test' @ branch spectre/kan-383-flow-add-a-file-flag-to-flow-state-set -->
<!-- predicted: the same command after this task — the count plus the five new test functions, which add no t.Run subtests -->
**Build:** green
**Commit:** `feat(flow-cli): read state set's record from a -file path`

- [x] 2. Make `-file` the canonical form in `state-file.md`

  - [x] **Step 1: replace the usage block at the top of `skills/flow-contracts/state-file.md`**

  ```bash verified:current lines read against skills/flow-contracts/state-file.md's opening code block on this branch
  flow state get [-C dir] <name>               # prints the record's JSON to stdout
  flow state set [-C dir] [-file path] <name>  # reads the whole record as JSON from path, or from stdin when -file is absent
  ```

  - [x] **Step 2: replace the canonical form in "Read it, write it"**

  The section's code block becomes the single non-compound command, and its
  closing paragraph's "from stdin" becomes "from the file `-file` names, or
  from stdin when the flag is absent". Directly after that paragraph, add:

  ```text unverified:confirm the paragraph reads as contract prose, not as a new gate — check-contract-budget.sh and check-vocabulary.sh must both stay clean
  **The file form is the allowlist-friendly form.** Harness permission
  classifiers block compound piped and heredoc commands — exactly what the
  piped canonical form was. Writing the record to a scratch file first
  (`printf '%s' "$RECORD_JSON" > "$RECORD_FILE"` — an independent command,
  allowed on its own terms) and then running `flow state set -file
  "$RECORD_FILE" "$NAME" -C "$DIR"` makes the write itself one non-compound
  command that a `Bash(flow state set *)` allowlist rule covers cleanly
  (KAN-383). The piped form `printf '%s' "$RECORD_JSON" | flow state set
  "$NAME" -C "$DIR"` remains supported as the stdin alternative.
  ```

  - [x] **Step 3: extend the local-input-error sentence in "The pipeline never blocks"**

  The paragraph beginning "A CLI usage error" gains the unreadable file
  alongside the sources it already names:

  ```text verified:current paragraph read against skills/flow-contracts/state-file.md's "The pipeline never blocks" section on this branch
  A CLI usage error — non-JSON stdin, a `-file` path that cannot be read, JSON
  that is not an object, a stdin or `-file` payload over the CLI's own size
  cap, or a `worktrees` value that is neither `null` nor a sha (**The record**
  below) — is a local input error, not a store failure: it is reported to
  stderr and exits 2, never falls back, and never reaches the network.
  ```

  - [x] **Step 4: run this task's verification**

  ```bash verified:commands as named in .flow/project.md's ## lint list — the entries whose corpus covers skills/ markdown
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  scripts/check-markdown-integrity.py
  ```

  Expected: all four exit 0. `state-file.md`'s byte size stays inside its
  declared budget row, so no budget edit is part of this task.

  - [x] **Step 5: commit**

  ```bash verified:subject scope names the module this task's own Files field carries, per commit-scope-is-the-module
  git add skills/flow-contracts/state-file.md
  git commit -m "docs(flow-contracts): make -file the canonical state set form"
  ```

**Files:** `skills/flow-contracts/state-file.md`
**Tests:** none — contract markdown only
**Regression:** none — the commit touches no Go code, so no test behaviour can change; reverting it restores the piped canonical form and leaves the CLI flag documented nowhere
**Baseline:** before=200 after=200
<!-- measured: cd stats && go test ./cmd/flow -count=1 -v | grep -c '^=== RUN   Test' @ branch spectre/kan-383-flow-add-a-file-flag-to-flow-state-set -->
<!-- predicted: the same command after this task — a markdown-only commit cannot move the Go test count -->
**Build:** green
**Commit:** `docs(flow-contracts): make -file the canonical state set form`

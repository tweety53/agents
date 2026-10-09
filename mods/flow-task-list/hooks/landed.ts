// Which plan tasks have landed on a change branch, read from git. No imports, so
// scripts/test-flow-task-list-landed.sh runs it under node against a real repository.

// The task numbers in the `Task-Id:` trailers `git log` printed, one value per line.
export const landed = (log: string): number[] => (log.match(/\d+/g) ?? []).map(Number)

export type Run = (argv: string[]) => Promise<{ exitCode: number; stdout: string }>

// The tasks whose `Task-Id:` commits are on the worktree `dir`'s branch since its base — the base recorded as
// `branch.<current>.flowBase` (its `refs/remotes/origin/` ref first), else `refs/remotes/origin/HEAD`, as
// scripts/check-task-records.py resolves it, offline; none when git cannot say. Only this one worktree is read:
// a task landing in a peer repo's worktree counts once it is ticked.
export const landedIn = async (run: Run, dir: string): Promise<number[]> => {
  const git = async (...args: string[]) => {
    const r = await run(['git', '-C', dir, ...args]).catch(() => null)
    return r?.exitCode === 0 ? r.stdout.trim() : null
  }
  const cur = await git('branch', '--show-current')
  const rec = cur ? await git('config', '--get', `branch.${cur}.flowBase`) : null
  const base = !rec
    ? 'refs/remotes/origin/HEAD'
    : (await git('rev-parse', '--verify', '--quiet', `refs/remotes/origin/${rec}`)) !== null
      ? `refs/remotes/origin/${rec}`
      : rec
  const log = await git('log', '--format=%(trailers:key=Task-Id,valueonly)', `${base}..HEAD`)
  return log === null ? [] : landed(log)
}

// Small formatting helpers shared by the eight views. None of them ever
// substitutes a default for a missing value -- that is exactly what
// Unavailable.tsx exists to prevent -- they format only the value they are
// actually given.
export function formatUsd(value: number): string {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 4 }).format(
    value,
  );
}

export function formatInt(value: number): string {
  return new Intl.NumberFormat("en-US").format(Math.round(value));
}

export function formatRatio(value: number): string {
  return value.toFixed(2);
}

export function formatMs(value: number): string {
  if (value < 1000) return `${Math.round(value)} ms`;
  const seconds = value / 1000;
  if (seconds < 60) return `${seconds.toFixed(1)} s`;
  const minutes = seconds / 60;
  return `${minutes.toFixed(1)} min`;
}

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString();
}

/**
 * Formats a stage run's wall-clock duration, or reports that it is still
 * running when it has no end instant yet. Deliberately distinct from
 * Unavailable's "unmeasured" reading: an open run's duration is not an
 * absent measurement, it simply hasn't happened yet, and rendering it
 * through Unavailable would blur that with a run whose duration truly
 * went unrecorded (StageRunTable.tsx, StageTimeline.tsx -- task 18's own
 * "a run with no endedAt is open, not zero-length" requirement).
 */
export function formatDurationOrOpen(startedAt: string, endedAt: string | undefined): string {
  if (!endedAt) return "still running";
  return formatMs(new Date(endedAt).getTime() - new Date(startedAt).getTime());
}

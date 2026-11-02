// Chip is the KAN-1048 filter chip: a toggle button styled by `lb-fchip`,
// pressed when its value is the group's selection. Shared by the live state
// board and the stage leaderboard (decision leaderboard-shared-chip).
// ChipGroup is the labelled `lb-fgroup` a row of chips sits in, shared by the
// chip-bar views (decision shared-chip-group-and-week-window).
import type { ReactNode } from "react";

export function Chip({ on, onClick, title, children }: { on: boolean; onClick: () => void; title?: string; children: ReactNode }) {
  return (
    <button type="button" className="lb-fchip" aria-pressed={on} onClick={onClick} title={title}>
      {children}
    </button>
  );
}

export function ChipGroup({ name, children }: { name: string; children: ReactNode }) {
  return (
    <span className="lb-fgroup" role="group" aria-label={name}>
      <span className="lb-fl">{name}</span>
      {children}
    </span>
  );
}

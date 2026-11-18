import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";

const css = readFileSync(`${process.cwd()}/src/theme.css`, "utf8");

// The body of the first `{ ... }` block opened at or after `start`, nested braces included.
function block(start: number): string {
  const open = css.indexOf("{", start);
  let depth = 0;
  for (let i = open; i < css.length; i++) {
    if (css[i] === "{") depth++;
    if (css[i] === "}" && --depth === 0) return css.slice(open + 1, i);
  }
  throw new Error("unclosed block");
}

const count = (body: string, decl: string) => body.split(decl).length - 1;

describe("theme.css", () => {
  it("forced-light theme declares the light marker and heat", () => {
    const forced = block(css.indexOf(':root[data-theme="light"]'));
    const media = block(css.indexOf("@media (prefers-color-scheme: light)"));
    for (const decl of ["--marker: #8250df;", "--heat: 31, 111, 235;"]) {
      expect(count(forced, decl)).toBe(1);
      expect(count(media, decl)).toBe(1);
    }
  });
});

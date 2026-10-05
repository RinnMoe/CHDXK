import { describe, expect, it } from "vitest"
import { buildThemeCSS, defaultTheme, parseTheme } from "./site-theme"

describe("site theme", () => {
  it("builds both modes from validated colors", () => {
    const value = structuredClone(defaultTheme)
    value.light.primary = "#123456"
    expect(buildThemeCSS(value)).toContain(":root{")
    expect(buildThemeCSS(value)).toContain("--primary:#123456;")
    expect(buildThemeCSS(value)).toContain(".dark{")
  })
  it("rejects missing and unrecognized tokens and CSS injection", () => {
    for (const value of [
      { light: {} },
      { ...defaultTheme, extra: {} },
      {
        ...defaultTheme,
        light: { ...defaultTheme.light, primary: "red;display:none" },
      },
    ])
      expect(() => parseTheme(value)).toThrow()
  })
})

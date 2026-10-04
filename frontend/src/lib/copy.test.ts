import { afterEach, describe, expect, it } from "vitest"
import { copy, copyDirectory, setCopyOverrides } from "./copy"

afterEach(() => setCopyOverrides({}))

describe("site copy", () => {
  it("uses defaults and substitutes dynamic values without evaluating markup", () => {
    const key = Object.keys(copyDirectory).find((key) =>
      copyDirectory[key].text.includes("{v0}")
    )!
    setCopyOverrides({ [key]: "数量 {v0}，内容 {v1}" })
    expect(copy(key, { v0: 12, v1: "<script>alert(1)</script>" })).toBe(
      "数量 12，内容 <script>alert(1)</script>"
    )
    expect(copy(key)).toBe("数量 ，内容 ")
  })
  it("keeps whitespace and falls back for missing overrides", () => {
    const key = "copy.user.config.brand.743e9916a4b2"
    expect(copy(key)).toBe(copyDirectory[key].text)
    setCopyOverrides({ [key]: "  新站点名称  " })
    expect(copy(key)).toBe("  新站点名称  ")
    expect(copy("unknown.key")).toBe("unknown.key")
  })
})

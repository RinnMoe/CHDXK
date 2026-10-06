import { renderToStaticMarkup } from "react-dom/server"
import { describe, expect, it } from "vitest"
import { SchoolLoginFeedback } from "./school-login-feedback"

describe("school login progress", () => {
  it("reports confirmation, school connection, and semester discovery separately", () => {
    const confirming = renderToStaticMarkup(
      <SchoolLoginFeedback state="authenticating" />
    )
    expect(confirming.match(/aria-label="已完成"/g)).toHaveLength(1)
    expect(confirming).toContain('aria-current="step"')
    expect(confirming).toContain("无需重复扫码")
    const discovering = renderToStaticMarkup(
      <SchoolLoginFeedback state="discovering" />
    )
    expect(discovering.match(/aria-label="已完成"/g)).toHaveLength(2)
    const ready = renderToStaticMarkup(<SchoolLoginFeedback state="ready" />)
    expect(ready.match(/aria-label="已完成"/g)).toHaveLength(3)
    expect(ready).not.toContain('aria-label="进行中"')
  })
  it("does not imply success while waiting or after failure", () => {
    const waiting = renderToStaticMarkup(
      <SchoolLoginFeedback state="waiting" />
    )
    expect(waiting).not.toContain('aria-label="已完成"')
    expect(renderToStaticMarkup(<SchoolLoginFeedback state="failed" />)).toBe(
      ""
    )
  })
})

import { expect, it, vi } from "vitest"
import { completeQrLogin, parseWechatPage, qrLogin } from "./qr-login.js"

function callbackPage(finalService = true, unbound = false, direct = false) {
  let current = "about:blank"
  const urls: string[] = []
  const page = {
    goto: async (url: string) => {
      urls.push(url)
      current = (direct || (urls.length > 1 && finalService)) ? "http://bkjw.chd.edu.cn/eams/home.action" : "https://ids.chd.edu.cn/authserver/index"
      return {ok: () => true}
    },
    content: async () => unbound ? "首次使用微信联合登录" : "已登录",
    url: () => current,
  }
  return {page, urls}
}

it("exchanges CAS authentication for an EAMS session when callback stays on CAS", async () => {
  const {page, urls} = callbackPage()
  await completeQrLogin(page as never, new URL("https://ids.chd.edu.cn/authserver/callback?code=test"))
  expect(urls).toHaveLength(2)
  expect(new URL(urls[1]!).searchParams.get("service")).toBe("http://bkjw.chd.edu.cn/eams/login.action")
})

it("does not re-login when callback already reaches EAMS", async () => {
  const {page, urls} = callbackPage(true, false, true)
  await completeQrLogin(page as never, new URL("https://ids.chd.edu.cn/authserver/callback?code=test"))
  expect(urls).toHaveLength(1)
})

it("rejects an unbound WeChat account before attempting EAMS login", async () => {
  const {page, urls} = callbackPage(true, true)
  await expect(completeQrLogin(page as never, new URL("https://ids.chd.edu.cn/authserver/callback?code=test"))).rejects.toThrow("WECHAT_UNBOUND")
  expect(urls).toHaveLength(1)
})

it("does not report login success when CAS sends the user back to login", async () => {
  const {page} = callbackPage(false)
  await expect(completeQrLogin(page as never, new URL("https://ids.chd.edu.cn/authserver/callback?code=test"))).rejects.toThrow("AUTH_REQUIRED")
})

it("reads QR image and callback without executing login page scripts", () => {
  const result = parseWechatPage('<html><body><img class="web_qrcode_img" src="/connect/qrcode/test"><script>var t = "https://ids.chd.edu.cn/authserver/callback?state=abc&amp;service=test"; var poll="qrconnect?uuid=abcdefghijkl";</script></body></html>', "https://open.weixin.qq.com/connect/qrconnect")
  expect(result.uuid).toBe("abcdefghijkl")
  expect(result.imageUrl.href).toBe("https://open.weixin.qq.com/connect/qrcode/test")
  expect(result.callbackUrl.searchParams.get("state")).toBe("abc")
})
it("rejects a callback outside school CAS", () => {
  expect(() => parseWechatPage('<img class="web_qrcode_img" src="/qrcode"><script>var t="https://example.com/steal"; "qrconnect?uuid=abcdefghijkl"</script>', "https://open.weixin.qq.com/")).toThrow()
})

it.each([true, false])("reports phone confirmation before callback navigation (scan event: %s)", async scanned => {
  vi.useFakeTimers()
  try {
    const events: {type: string}[] = []
    let release: () => void = () => {}
    const gate = new Promise<void>(resolve => {release = resolve})
    const {page} = callbackPage(true, false, true)
    const goto = page.goto
    let closed = false
    const callback = {...page, goto: async (url: string) => {await gate; return goto(url)}, close: async () => {closed = true}}
    let polls = 0
    const context = {
      request: {
        storageState: async () => ({cookies: []}),
        get: async (url: string) => {
          const path = new URL(url).pathname
          const html = path === "/authserver/login" ? '<iframe id="iframe" src="/authserver/wxEntry"></iframe>'
            : path === "/connect/qrconnect" ? '<img class="web_qrcode_img" src="/connect/qrcode/test"><script>var t="https://ids.chd.edu.cn/authserver/callback"; "qrconnect?uuid=abcdefghijkl"</script>'
            : path === "/connect/l/qrconnect" ? (scanned && ++polls < 3 ? 'window.wx_errcode=404;' : 'window.wx_errcode=200;window.wx_code="fixture-code";') : ""
          return {text:async () => html, url:() => "https://open.weixin.qq.com/connect/qrconnect", ok:() => true, body:async () => Buffer.from("image")}
        },
      },
      newPage: async () => callback,
    }
    let completed = false
    const login = qrLogin(context as never, event => events.push(event as {type:string}), new AbortController().signal).then(() => {completed = true})
    await vi.advanceTimersByTimeAsync(6000)
    expect(events.map(event => event.type)).toEqual(scanned ? ["qr", "scanned", "authenticating"] : ["qr", "authenticating"])
    expect(completed).toBe(false)
    expect(JSON.stringify(events)).not.toContain("fixture-code")
    release()
    await login
    expect(closed).toBe(true)
  } finally {vi.useRealTimers()}
})

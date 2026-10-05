import { randomBytes } from "node:crypto"
import { parseHTML } from "linkedom"
import type { BrowserContext, Page } from "playwright"

const CAS = "https://ids.chd.edu.cn"
const SERVICE = "http://bkjw.chd.edu.cn/eams/login.action"
const loginUrl = `${CAS}/authserver/login?service=${encodeURIComponent(SERVICE)}`

export async function completeQrLogin(page: Page, callbackUrl: URL) {
  // Run the callback as a browser navigation: CAS may return a form/script,
  // and successful CAS authentication alone does not establish an EAMS session.
  const callback = await page.goto(callbackUrl.href, {waitUntil: "load", timeout: 30000})
  if (!callback?.ok()) throw new Error("QR_LOGIN_FAILED")
  if ((await page.content()).includes("首次使用微信联合登录")) throw new Error("WECHAT_UNBOUND")
  if (new URL(page.url()).hostname !== "bkjw.chd.edu.cn") {
    await page.goto(loginUrl, {waitUntil: "load", timeout: 30000})
  }
  if (new URL(page.url()).hostname !== "bkjw.chd.edu.cn") throw new Error("AUTH_REQUIRED")
}
function htmlDecode(value: string) {
  return value.replace(/&(?:[a-z]+|#\d+|#x[a-f\d]+);/gi, (entity) => parseHTML(`<html><body>${entity}</body></html>`).document.body.textContent ?? entity)
}
export function parseWechatPage(html: string, url: string) {
  const uuid = /qrconnect\?uuid=([a-zA-Z0-9_-]{10,})/.exec(html)?.[1]
  const image = parseHTML(html).document.querySelector("img.web_qrcode_img")?.getAttribute("src")
  const callback = /var\s+t\s*=\s*"([^"]+)"/.exec(html)?.[1]
    ?? /redirect_uri["']?\s*[=:]\s*["']([^"']+)["']/.exec(html)?.[1]
  if (!uuid || !image) throw new Error("QR_PAGE_CHANGED")
  const state = /state=([a-zA-Z0-9_-]+)/.exec(html)?.[1] ?? ""
  const callbackUrl = new URL(callback ? decodeURIComponent(htmlDecode(callback)) : `${CAS}/authserver/callback`)
  if (callbackUrl.origin !== CAS || !callbackUrl.pathname.startsWith("/authserver/")) throw new Error("QR_CALLBACK_INVALID")
  if (state && !callbackUrl.searchParams.has("state")) callbackUrl.searchParams.set("state", state)
  const imageUrl = new URL(image, url)
  if (imageUrl.origin !== "https://open.weixin.qq.com") throw new Error("QR_IMAGE_INVALID")
  return {uuid, state, callbackUrl, imageUrl}
}

export async function qrLogin(context: BrowserContext, emit: (event: object) => void, signal: AbortSignal) {
  const request = context.request
  const login = await request.get(loginUrl)
  const iframe = parseHTML(await login.text()).document.querySelector("iframe#iframe, section.auth-qroce iframe, iframe")?.getAttribute("src")
  if (!iframe) throw new Error("QR_ENTRY_CHANGED")
  const iframeUrl = new URL(iframe, CAS)
  if (iframeUrl.origin !== CAS) throw new Error("QR_ENTRY_INVALID")
  const entry = await request.get(iframeUrl.href)
  const entryText = htmlDecode(await entry.text()).replaceAll("\\/", "/")
  const wxUrl = entry.url().startsWith("https://open.weixin.qq.com/") ? entry.url() : /https:\/\/open\.weixin\.qq\.com[^\s"'<>]+/.exec(entryText)?.[0]
  if (!wxUrl) throw new Error("QR_ENTRY_CHANGED")
  const wxPage = await request.get(wxUrl, {headers: {Referer: iframeUrl.href}})
  const html = await wxPage.text()
  const qr = parseWechatPage(html, wxPage.url())
  const storage = await request.storageState()
  const muid = storage.cookies.find((cookie) => cookie.name === "yyb_muid")?.value
    ?? /(?:yyb_muid|muid)["']?\s*[=:]\s*["']([^"']+)["']/.exec(html)?.[1] ?? randomBytes(16).toString("hex")
  // APIRequestContext has no addCookies; the cookie is sent only to the WeChat poll host.
  const image = await request.get(qr.imageUrl.href)
  if (!image.ok()) throw new Error("QR_IMAGE_UNAVAILABLE")
  emit({type: "qr", qr: `data:image/png;base64,${(await image.body()).toString("base64")}`})
  const deadline = Date.now() + 180000
  let last = ""
  while (!signal.aborted && Date.now() < deadline) {
    await new Promise<void>((resolve) => setTimeout(resolve, 2000))
    if (signal.aborted) throw new Error("CANCELLED")
    const poll = new URL("https://lp.open.weixin.qq.com/connect/l/qrconnect")
    poll.searchParams.set("uuid", qr.uuid)
    if (last && last !== "408") poll.searchParams.set("last", last)
    let text: string
    try { text = await (await request.get(poll.href, {timeout: 10000, headers: {Referer: "https://open.weixin.qq.com/", Cookie: `yyb_muid=${muid}`}})).text() }
    catch { continue }
    const code = /window\.wx_code\s*=\s*['"]([^'"]+)['"]/.exec(text)?.[1]
    last = /window\.wx_errcode\s*=\s*(\d+)/.exec(text)?.[1] ?? last
    if (code) {
      qr.callbackUrl.searchParams.set("code", code)
      const page = await context.newPage()
      try { await completeQrLogin(page, qr.callbackUrl) }
      finally { await page.close() }
      return
    }
    if (last === "404") emit({type: "scanned"})
    if (["402", "403", "500"].includes(last)) break
  }
  throw new Error(signal.aborted ? "CANCELLED" : "QR_EXPIRED")
}

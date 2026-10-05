import { expect, it } from "vitest"
import { parseWechatPage } from "./qr-login.js"

it("reads QR image and callback without executing login page scripts", () => {
  const result = parseWechatPage('<html><body><img class="web_qrcode_img" src="/connect/qrcode/test"><script>var t = "https://ids.chd.edu.cn/authserver/callback?state=abc&amp;service=test"; var poll="qrconnect?uuid=abcdefghijkl";</script></body></html>', "https://open.weixin.qq.com/connect/qrconnect")
  expect(result.uuid).toBe("abcdefghijkl")
  expect(result.imageUrl.href).toBe("https://open.weixin.qq.com/connect/qrcode/test")
  expect(result.callbackUrl.searchParams.get("state")).toBe("abc")
})
it("rejects a callback outside school CAS", () => {
  expect(() => parseWechatPage('<img class="web_qrcode_img" src="/qrcode"><script>var t="https://example.com/steal"; "qrconnect?uuid=abcdefghijkl"</script>', "https://open.weixin.qq.com/")).toThrow()
})

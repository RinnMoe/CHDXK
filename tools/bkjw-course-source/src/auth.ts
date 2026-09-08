import { mkdtemp, rm } from "node:fs/promises"
import { spawn } from "node:child_process"
import { tmpdir } from "node:os"
import { dirname, join } from "node:path"

import type { Locator, Page } from "playwright"

import { BkjwError, PageStructureError } from "./types.js"

const CAPTCHA_PATTERN = /^[A-Za-z0-9]{4}$/
const CAPTCHA_ATTEMPTS = 3

export interface CredentialLoginOptions {
  username: string
  password: string
  promptCaptcha: (imagePath: string | null) => Promise<string>
}

/**
 * The CAS page is rendered from hidden templates, so selectors must be
 * resolved against the visible account-login form rather than a static id.
 */
export async function hasCasLoginForm(page: Page): Promise<boolean> {
  const username = await firstVisible(page, [
    ".login-main .m-account #username",
    ".login-main input[name='username']",
    "input[name='username']",
  ])
  const password = await firstVisible(page, [
    ".login-main .m-account #password",
    ".login-main input[type='password']",
    "input[type='password']",
  ])
  return username !== null && password !== null
}

export function isCasLoginUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return url.protocol === "https:" && url.hostname === "ids.chd.edu.cn" && url.pathname.startsWith("/authserver/login")
  } catch {
    return false
  }
}

export async function waitForCasLoginForm(page: Page, timeout = 10_000): Promise<boolean> {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    if (!isCasLoginUrl(page.url())) return false
    if (await hasCasLoginForm(page)) return true
    await page.waitForTimeout(200)
  }
  return hasCasLoginForm(page)
}

export async function loginWithCredentials(
  page: Page,
  options: CredentialLoginOptions,
): Promise<void> {
  const usernameInput = await firstVisible(page, [
    ".login-main .m-account #username",
    ".login-main input[name='username']",
    "input[name='username']",
  ])
  const passwordInput = await firstVisible(page, [
    ".login-main .m-account #password",
    ".login-main input[type='password']",
    "input[type='password']",
  ])
  if (!usernameInput || !passwordInput) {
    throw new PageStructureError("统一身份认证页缺少可见的账号或密码输入框")
  }

  const username = options.username.trim()
  if (!username) throw new BkjwError("账号不能为空", "LOGIN_INPUT")
  if (!options.password) throw new BkjwError("密码不能为空", "LOGIN_INPUT")

  for (let attempt = 1; attempt <= CAPTCHA_ATTEMPTS; attempt += 1) {
    // A failed CAS submission can replace the form with a fresh document.
    // Refill both fields before every retry so the password is never assumed
    // to have survived a navigation.
    await usernameInput.fill(username)
    await usernameInput.press("Tab").catch(() => undefined)
    await page.waitForTimeout(300)
    await passwordInput.fill(options.password)
    await enableRememberMe(page)

    const captchaNeeded = await checkNeedCaptcha(page, username)
    const captchaInput = await firstVisible(page, [
      ".login-main .m-account #captcha",
      ".login-main input[name='captcha']",
      "input[name='captcha']:visible",
    ])
    const captchaVisible = captchaInput !== null || (await hasVisibleCaptchaImage(page))

    if (captchaNeeded || captchaVisible) {
      const imagePath = await prepareCaptcha(page)
      try {
        const visibleCaptchaInput = await firstVisible(page, [
          ".login-main .m-account #captcha",
          ".login-main input[name='captcha']",
          "input[name='captcha']:visible",
        ])
        if (!visibleCaptchaInput) throw new PageStructureError("统一身份认证页要求验证码，但缺少验证码输入框")
        const captcha = await promptValidCaptcha(options.promptCaptcha, imagePath)
        await visibleCaptchaInput.fill(captcha)
      } finally {
        await removeCaptchaImage(imagePath)
      }
    }

    await submitLoginForm(page)
    if (!(await hasCasLoginForm(page))) return

    const errorKind = await loginErrorKind(page)
    const captchaAppeared = await hasVisibleCaptchaImage(page)
    if (captchaAppeared || errorKind === "captcha") {
      if (attempt < CAPTCHA_ATTEMPTS) {
        await page.waitForTimeout(300)
        continue
      }
      throw new BkjwError("验证码校验失败，请重新运行 login", "CAPTCHA_FAILED")
    }

    throw new BkjwError(errorKind === "credentials" ? "账号或密码错误" : "统一身份认证登录失败", "LOGIN_FAILED")
  }

  throw new BkjwError("统一身份认证登录失败", "LOGIN_FAILED")
}

async function firstVisible(page: Page, selectors: string[]): Promise<Locator | null> {
  for (const selector of selectors) {
    const locator = page.locator(selector)
    const count = await locator.count()
    for (let index = 0; index < count; index += 1) {
      const candidate = locator.nth(index)
      if (await candidate.isVisible().catch(() => false)) return candidate
    }
  }
  return null
}

async function checkNeedCaptcha(page: Page, username: string): Promise<boolean> {
  try {
    const result = await page.evaluate(async (value) => {
      const endpoint = new URL("checkNeedCaptcha.htl", window.location.href)
      endpoint.searchParams.set("username", value)
      endpoint.searchParams.set("_", String(Date.now()))
      const response = await fetch(endpoint, {
        credentials: "include",
        headers: { Accept: "application/json" },
      })
      if (!response.ok) return null
      const data = (await response.json()) as { isNeed?: unknown }
      return typeof data.isNeed === "boolean" ? data.isNeed : null
    }, username)
    if (result !== null) return result
  } catch {
    // The visible form remains the safe fallback when the helper endpoint changes.
  }
  return hasVisibleCaptchaImage(page)
}

async function enableRememberMe(page: Page): Promise<void> {
  const checkbox = await firstVisible(page, [
    ".login-main .m-account #rememberMe",
    ".login-main input[name='rememberMe']",
  ])
  if (!checkbox || (await checkbox.isChecked().catch(() => true))) return
  await checkbox.check().catch(() => undefined)
}

async function hasVisibleCaptchaImage(page: Page): Promise<boolean> {
  const image = await firstVisible(page, [
    ".login-main .m-account #captchaImg",
    ".login-main img[id='captchaImg']",
  ])
  if (!image) return false
  const src = await image.getAttribute("src")
  return Boolean(src)
}

async function prepareCaptcha(page: Page): Promise<string | null> {
  await page.evaluate(() => {
    // The current CAS theme may advertise a slider (`captchaSwitch=2`), while
    // the same authenticated endpoint still supports the classic image
    // challenge used by CHUAuthSDK. Force that form so the user can inspect
    // the image and enter its value without putting credentials in a request.
    const globals = globalThis as { captchaSwitch?: string; needCaptcha?: boolean }
    globals.captchaSwitch = "1"
    globals.needCaptcha = true
    const root = document.querySelector(".login-main .m-account")
    const image = root?.querySelector("#captchaImg") as HTMLImageElement | null
    const input = root?.querySelector("#captcha") as HTMLInputElement | null
    const container = root?.querySelector("#captchaDiv") as HTMLElement | null
    if (container) {
      container.classList.remove("hide")
      container.style.display = "block"
    }
    if (input) input.style.display = "inline-block"
    if (image) {
      image.style.display = "inline-block"
      image.src = `/authserver/getCaptcha.htl?_=${Date.now()}`
    }
  })
  await page.waitForTimeout(250)

  const image = await firstVisible(page, [
    ".login-main .m-account #captchaImg",
    ".login-main img[id='captchaImg']",
  ])
  if (!image) return null

  try {
    await image.waitFor({ state: "visible", timeout: 5_000 })
    await page
      .waitForFunction(
        () => {
          const candidate = document.querySelector(".login-main .m-account #captchaImg") as HTMLImageElement | null
          return Boolean(candidate?.complete && candidate.naturalWidth > 0)
        },
        undefined,
        { timeout: 5_000 },
      )
      .catch(() => undefined)
    const directory = await mkdtemp(join(tmpdir(), "bkjw-captcha-"))
    const imagePath = join(directory, "captcha.png")
    await image.screenshot({ path: imagePath })
    openCaptchaImage(imagePath)
    return imagePath
  } catch {
    return null
  }
}

async function promptValidCaptcha(
  promptCaptcha: CredentialLoginOptions["promptCaptcha"],
  imagePath: string | null,
): Promise<string> {
  const value = (await promptCaptcha(imagePath)).trim()
  if (!CAPTCHA_PATTERN.test(value)) throw new BkjwError("验证码格式不正确，应为 4 位字母或数字", "CAPTCHA_INVALID")
  return value
}

async function removeCaptchaImage(imagePath: string | null): Promise<void> {
  if (!imagePath) return
  await rm(dirname(imagePath), { recursive: true, force: true }).catch(() => undefined)
}

function openCaptchaImage(imagePath: string): void {
  try {
    if (process.platform === "win32") {
      const child = spawn("rundll32.exe", ["url.dll,FileProtocolHandler", imagePath], {
        detached: true,
        stdio: "ignore",
        windowsHide: true,
      })
      child.unref()
    } else if (process.platform === "darwin") {
      const child = spawn("open", [imagePath], { detached: true, stdio: "ignore" })
      child.unref()
    } else {
      const child = spawn("xdg-open", [imagePath], { detached: true, stdio: "ignore" })
      child.unref()
    }
  } catch {
    // The browser still displays the captcha when no system image viewer exists.
  }
}

async function submitLoginForm(page: Page): Promise<void> {
  const submit = await firstVisible(page, [
    ".login-main #login_submit",
    ".login-main button[type='submit']",
    ".login-main input[type='submit']",
    "button[type='submit']",
    "input[type='submit']",
  ])
  if (!submit) throw new PageStructureError("统一身份认证页缺少登录按钮")

  const previousUrl = page.url()
  const navigation = page
    .waitForURL((url) => url.toString() !== previousUrl, { timeout: 12_000 })
    .catch(() => undefined)
  await submit.click({ timeout: 10_000 })
  await navigation
  await page.waitForLoadState("domcontentloaded", { timeout: 10_000 }).catch(() => undefined)
  await page.waitForTimeout(400)
}

async function loginErrorKind(page: Page): Promise<"captcha" | "credentials" | "other" | null> {
  const texts = await page
    .locator(
      ".login-main #showErrorTip:visible, .login-main #captchaErrorTip:visible, .login-main .item-error-tip:visible",
    )
    .allTextContents()
    .catch(() => [])
  const text = texts.join(" ").replace(/\s+/g, " ").trim()
  if (!text) return null
  if (/验证码|captcha/i.test(text)) return "captcha"
  if (/用户名|账号|密码|登录失败|认证失败/i.test(text)) return "credentials"
  return "other"
}

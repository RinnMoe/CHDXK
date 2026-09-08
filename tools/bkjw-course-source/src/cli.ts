import { createInterface } from "node:readline"
import { resolve } from "node:path"
import { pathToFileURL } from "node:url"

import { isCasLoginUrl, loginWithCredentials, waitForCasLoginForm } from "./auth.js"
import { BkjwClient, launchBkjwContext } from "./client.js"
import { fetchSemester } from "./orchestrator.js"
import { DEFAULT_OUTPUT_ROOT, DEFAULT_PROFILE_DIR } from "./storage.js"
import { AuthenticationRequiredError, InterruptedError, type SemesterOption } from "./types.js"

interface CliOptions {
  command: string
  semester?: string
  all: boolean
  output: string
  profile: string
  headless: boolean
  resume: boolean
  pageSize: number
  detailConcurrency: number
  help: boolean
}

const HELP = `长大教务全校开课数据采集器

用法：
  pnpm start login [--profile <目录>]
  pnpm start semesters [--headless]
  pnpm start fetch --semester <学期名称或站点 ID> [选项]
  pnpm start fetch --all [选项]

选项：
  --output <目录>              数据根目录（默认 data/bkjw）
  --profile <目录>             持久化浏览器会话目录
  --headless                   无头运行（默认可见；login 始终使用可见浏览器）
  --resume                     从该学期的 checkpoint.json 继续
  --page-size <数量>           请求的单页数量，按站点允许值向下取整（默认 1000）
  --detail-concurrency <数量>  详情页并发数，范围 1-8（默认 3）
  --all                        fetch 命令采集所有可发现学期
  --help                       显示帮助

首次使用：
  pnpm start login
  # 按终端提示输入账号密码；需要验证码时查看浏览器图片并手动输入
`

export async function runCli(argv: string[] = process.argv.slice(2)): Promise<number> {
  let options: CliOptions
  try {
    options = parseArgs(argv)
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error))
    console.error(HELP)
    return 2
  }
  if (options.help || !options.command) {
    console.log(HELP)
    return options.help ? 0 : 2
  }

  if (!["login", "semesters", "fetch"].includes(options.command)) {
    console.error(`未知命令：${options.command}`)
    console.error(HELP)
    return 2
  }
  if (options.command === "fetch" && !options.all && !options.semester) {
    console.error("fetch 需要 --semester <学期名称或站点 ID>，或使用 --all")
    return 2
  }
  if (options.command === "fetch" && options.all && options.semester) {
    console.error("fetch 不能同时使用 --all 和 --semester")
    return 2
  }

  let context
  try {
    context = await launchBkjwContext({
      profileDir: options.profile,
      headless: options.command === "login" ? false : options.headless,
    })
    const client = new BkjwClient(context)
    const page = await client.newPage()

    if (options.command === "login") {
      await client.navigateLogin(page)
      if (await waitForCasLoginForm(page)) {
        const username = await promptLine("请输入学工号/手机号：")
        const password = await promptSecret("请输入密码：")
        await loginWithCredentials(page, {
          username,
          password,
          promptCaptcha: (imagePath) => promptCaptcha(imagePath),
        })
      } else if (isCasLoginUrl(page.url())) {
        console.log("未识别统一身份认证的账号密码表单；请在浏览器中完成登录，完成后回终端按回车。")
        await waitForEnter()
      }
      await verifySavedLogin(client)
      await client.persistSessionCookies()
      console.log(`登录状态已验证并保存到：${options.profile}`)
      return 0
    }

    const semesters = await client.discoverSemesters(page)
    if (options.command === "semesters") {
      for (const semester of semesters) {
        console.log(`${semester.normalized}\t${semester.id}\t${semester.label}`)
      }
      return 0
    }

    if (semesters.length === 0) throw new Error("教务系统没有返回可用学期")
    const selected = options.all ? semesters : [resolveSemester(semesters, options.semester ?? "")]
    const controller = new AbortController()
    const onSignal = () => controller.abort()
    process.once("SIGINT", onSignal)
    let failed = false
    try {
      for (const semester of selected) {
        try {
          await fetchSemester(client, semester, {
            outputRoot: options.output,
            pageSize: options.pageSize,
            detailConcurrency: options.detailConcurrency,
            resume: options.resume,
            signal: controller.signal,
            log: (message) => console.log(`[${semester.normalized}] ${message}`),
          })
        } catch (error) {
          failed = true
          console.error(`[${semester.normalized}] ${error instanceof Error ? error.message : String(error)}`)
          if (error instanceof AuthenticationRequiredError || error instanceof InterruptedError) break
        }
      }
    } finally {
      process.removeListener("SIGINT", onSignal)
    }
    return failed ? 1 : 0
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error))
    return 1
  } finally {
    await context?.close()
  }
}

function parseArgs(argv: string[]): CliOptions {
  const helpBeforeCommand = argv[0] === "--help" || argv[0] === "-h"
  const command = helpBeforeCommand ? "" : (argv[0] ?? "")
  const options: CliOptions = {
    command,
    all: false,
    output: DEFAULT_OUTPUT_ROOT,
    profile: DEFAULT_PROFILE_DIR,
    headless: false,
    resume: false,
    pageSize: 1000,
    detailConcurrency: 3,
    help: helpBeforeCommand,
  }
  for (let index = 1; index < argv.length; index += 1) {
    const token = argv[index] ?? ""
    if (token === "--help" || token === "-h") {
      options.help = true
      continue
    }
    if (token === "--all") {
      options.all = true
      continue
    }
    if (token === "--headless") {
      options.headless = true
      continue
    }
    if (token === "--resume") {
      options.resume = true
      continue
    }
    const [name, inlineValue] = token.split("=", 2)
    const value = inlineValue ?? argv[++index]
    if (!value) throw new Error(`选项 ${name} 缺少值`)
    if (name === "--semester") options.semester = value
    else if (name === "--output") options.output = resolve(value)
    else if (name === "--profile") options.profile = resolve(value)
    else if (name === "--page-size") options.pageSize = positiveInteger(value, "--page-size")
    else if (name === "--detail-concurrency") options.detailConcurrency = positiveInteger(value, "--detail-concurrency")
    else throw new Error(`未知选项：${name}`)
  }
  if (options.detailConcurrency > 8) throw new Error("--detail-concurrency 不能大于 8")
  return options
}

function positiveInteger(value: string, name: string): number {
  const parsed = Number.parseInt(value, 10)
  if (!/^\d+$/.test(value) || parsed < 1) throw new Error(`${name} 必须是正整数`)
  return parsed
}

function resolveSemester(options: SemesterOption[], requested: string): SemesterOption {
  const normalized = requested.trim()
  const match = options.find(
    (option) => option.id === normalized || option.normalized === normalized || option.label === normalized,
  )
  if (!match) {
    const available = options.map((option) => `${option.normalized}(${option.id})`).join(", ")
    throw new Error(`找不到学期 ${requested}；可用学期：${available || "无"}`)
  }
  return match
}

function waitForEnter(): Promise<void> {
  return new Promise((resolvePromise) => {
    const readline = createInterface({ input: process.stdin, output: process.stdout })
    readline.question("按回车继续。", () => {
      readline.close()
      resolvePromise()
    })
  })
}

function promptLine(prompt: string): Promise<string> {
  return new Promise((resolvePromise) => {
    const readline = createInterface({ input: process.stdin, output: process.stdout })
    readline.question(prompt, (answer) => {
      readline.close()
      resolvePromise(answer)
    })
  })
}

function promptSecret(prompt: string): Promise<string> {
  if (!process.stdin.isTTY || !process.stdout.isTTY) return promptLine(prompt)

  return new Promise((resolvePromise, rejectPromise) => {
    const input = process.stdin
    const wasRaw = input.isRaw
    let value = ""
    const cleanup = () => {
      input.off("data", onData)
      input.setRawMode?.(Boolean(wasRaw))
      input.pause()
      process.stdout.write("\n")
    }
    const onData = (chunk: Buffer | string) => {
      for (const character of String(chunk)) {
        if (character === "\u0003") {
          cleanup()
          rejectPromise(new Error("已取消密码输入"))
          return
        }
        if (character === "\r" || character === "\n") {
          cleanup()
          resolvePromise(value)
          return
        }
        if (character === "\u007f" || character === "\b") {
          value = value.slice(0, -1)
          continue
        }
        value += character
      }
    }
    process.stdout.write(prompt)
    input.setRawMode?.(true)
    input.resume()
    input.on("data", onData)
  })
}

function promptCaptcha(imagePath: string | null): Promise<string> {
  if (imagePath) {
    console.log(`验证码图片已在浏览器中显示，并已打开临时图片：${imagePath}`)
  } else {
    console.log("验证码图片已在浏览器中显示；未能创建临时图片副本。")
  }
  return promptLine("请输入验证码（4 位字母或数字）：")
}

async function verifySavedLogin(client: BkjwClient): Promise<void> {
  let lastError: unknown
  for (let attempt = 1; attempt <= 5; attempt += 1) {
    const probe = await client.newPage()
    try {
      await client.navigateEntry(probe)
      return
    } catch (error) {
      lastError = error
      if (!(error instanceof AuthenticationRequiredError) || attempt === 5) break
      await probe.waitForTimeout(1_000)
    } finally {
      await client.closePage(probe)
    }
  }
  if (lastError instanceof AuthenticationRequiredError) {
    throw new AuthenticationRequiredError(
      "账号密码登录未建立可复用的教务会话；请确认登录已回跳到教务系统后再重试",
    )
  }
  throw lastError instanceof Error ? lastError : new Error("登录验证失败")
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  runCli().then((exitCode) => {
    process.exitCode = exitCode
  })
}

import fs from "node:fs"
import path from "node:path"
import { createHash } from "node:crypto"
import { parse } from "@babel/parser"

const root = path.resolve(import.meta.dirname, "..")
const catalogPath = path.join(root, "src/config/copy-catalog.json")
const catalog = fs.existsSync(catalogPath) ? JSON.parse(fs.readFileSync(catalogPath, "utf8")) : {}
const hash = (text) => createHash("sha256").update(text).digest("hex").slice(0, 12)
const han = /[\u3400-\u9fff]/
const english = new Set(["Close", "Next slide", "Previous slide", "More pages", "Toggle Sidebar", "Search", "Search...", "Loading..."])
const extract = process.argv.includes("--extract")
const check = process.argv.includes("--check")
let missing = 0

function files(dir) {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => entry.isDirectory() ? files(path.join(dir, entry.name)) : [path.join(dir, entry.name)])
}
function children(node) {
  return Object.entries(node).filter(([key]) => !["loc", "extra", "comments", "tokens"].includes(key)).flatMap(([, value]) => Array.isArray(value) ? value.filter((v) => v?.type) : value?.type ? [value] : [])
}
function jsxText(value) {
  const lines = value.replace(/\r/g, "").split("\n")
  let last = -1
  lines.forEach((line, i) => { if (line.trim()) last = i })
  return lines.map((line, i) => {
    let text = line.replace(/\t/g, " ")
    if (i > 0) text = text.replace(/^ +/, "")
    if (i < lines.length - 1) text = text.replace(/ +$/, "")
    return text && i !== last ? text + " " : text
  }).join("")
}

for (const file of files(path.join(root, "src"))) {
  const relative = path.relative(path.join(root, "src"), file).replaceAll("\\", "/")
  if (!/\.tsx?$/.test(file) || /(^mocks\/|\.test\.|^lib\/copy\.ts$)/.test(relative)) continue
  const source = fs.readFileSync(file, "utf8")
  const ast = parse(source, { sourceType: "module", plugins: ["typescript", "jsx"] })
  const edits = []
  const scope = relative.replace(/\.[^.]+$/, "").replaceAll("/", ".")
  const audience = relative.includes("admin/") || relative.startsWith("pages/admin") ? "admin" : "user"
  function visit(node, parent) {
    if (node.type === "CallExpression" && node.callee.name === "copy") {
      const key = node.arguments[0]?.value
      if (!catalog[key]) { console.error(`Missing catalog key: ${key}`); missing++ }
      else if (extract) catalog[key].location = relative
      return
    }
    let text, params = "", wrap = false
    if (node.type === "JSXText") { text = jsxText(node.value); wrap = true }
    if (node.type === "StringLiteral") {
      if (parent?.type === "TSLiteralType" || parent?.type === "BinaryExpression" || parent?.type === "ImportDeclaration" || (parent?.type === "ObjectProperty" && (parent.key === node || parent.key.name === "value"))) return
      text = node.value
      wrap = parent?.type === "JSXAttribute"
    }
    if (node.type === "TemplateLiteral" && node.quasis.some((q) => han.test(q.value.cooked ?? ""))) {
      text = node.quasis.map((q, i) => (q.value.cooked ?? q.value.raw) + (i < node.expressions.length ? `{v${i}}` : "")).join("")
      params = node.expressions.length ? ", { " + node.expressions.map((expression, i) => `v${i}: (${source.slice(expression.start, expression.end)})`).join(", ") + " }" : ""
    }
    const presentation = node.type === "JSXText" || (parent?.type === "JSXAttribute" && ["aria-label", "title", "placeholder", "alt"].includes(parent.name.name))
    if (text?.trim() && (han.test(text) || english.has(text) || (presentation && /[a-zA-Z]/.test(text)))) {
      // Date format tokens and identity/data sentinels are protocol values, not prose.
      if (/^[YMDHmsd年/月日时分秒\s:.-]+$/.test(text)) return
      if (!extract) { console.error(`Uncatalogued copy: ${relative}:${node.loc.start.line}`); missing++; return }
      const key = `copy.${audience}.${scope}.${hash(text)}`
      catalog[key] ??= { text, audience, location: relative }
      const expression = `copy(${JSON.stringify(key)}${params})`
      edits.push({ start: node.start, end: node.end, value: wrap ? `{${expression}}` : expression })
      return
    }
    for (const child of children(node)) visit(child, node)
  }
  visit(ast.program)
  if (edits.length) {
    let result = source
    for (const edit of edits.sort((a, b) => b.start - a.start)) result = result.slice(0, edit.start) + edit.value + result.slice(edit.end)
    const existing = /import \{ ([^\n]+) \} from "@\/lib\/copy"/.exec(result)
    if (existing) {
      if (!existing[1].split(",").some((name) => name.trim() === "copy")) result = result.replace(existing[0], `import { copy, ${existing[1]} } from "@/lib/copy"`)
    } else if (result.startsWith('"use client"')) result = result.replace('"use client"', '"use client"\n\nimport { copy } from "@/lib/copy"')
    else result = 'import { copy } from "@/lib/copy"\n\n' + result
    fs.writeFileSync(file, result, "utf8")
  }
}

// Backend messages remain stable identifiers at their source; presentation uses this directory.
for (const file of files(path.resolve(root, "../backend"))) {
  if (!file.endsWith(".go") || file.endsWith("_test.go") || file.includes(`${path.sep}sitecopy${path.sep}`)) continue
  const relative = path.relative(path.resolve(root, "../backend"), file).replaceAll("\\", "/")
  if (!/^(pkg\/apperr\/|internal\/interface\/web\/|internal\/app\/site_settings|internal\/application\/|internal\/domain\/[^/]+\/(errors|reshub)\.go)/.test(relative)) continue
  const source = fs.readFileSync(file, "utf8")
  for (const match of source.matchAll(/"((?:[^"\\]|\\.)*)"/g)) {
    let text
    try { text = JSON.parse(match[0]) } catch { continue }
    if (!han.test(text) || ["学年", "学期"].includes(text)) continue
    const key = `copy.server.${hash(text)}`
    catalog[key] ??= { text, audience: /API Key|审计|系统设置|导入|教学班|客户端/.test(text) ? "technical" : "user", location: relative }
  }
}
const serialized = JSON.stringify(Object.fromEntries(Object.entries(catalog).sort(([a], [b]) => a.localeCompare(b))), null, 2) + "\n"
const backendPath = path.resolve(root, "../backend/pkg/sitecopy/catalog.json")
if (check) {
  if (fs.readFileSync(catalogPath, "utf8") !== serialized || fs.readFileSync(backendPath, "utf8") !== serialized) { console.error("Copy catalogs are out of sync"); missing++ }
} else {
  fs.mkdirSync(path.dirname(backendPath), { recursive: true })
  fs.writeFileSync(catalogPath, serialized, "utf8")
  fs.writeFileSync(backendPath, serialized, "utf8")
}
console.log(`${Object.values(catalog).filter((entry) => entry.editable).length} editable entries; ${Object.keys(catalog).length} catalog entries; ${missing} uncatalogued items`)
if (missing) process.exitCode = 1

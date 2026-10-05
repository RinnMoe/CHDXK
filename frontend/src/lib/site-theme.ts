import defaults from "@/config/theme-defaults.json"

export type ThemeMode = keyof typeof defaults
export type ThemeToken = keyof typeof defaults.light
export type ThemePalette = Record<ThemeMode, Record<ThemeToken, string>>
export const defaultTheme: ThemePalette = defaults
export const themeSettingKey = "theme.palette"
const colorPattern = /^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$/

export function parseTheme(value: unknown): ThemePalette {
  if (!value || typeof value !== "object" || Object.keys(value).length !== 2)
    throw new Error("Invalid theme")
  for (const mode of ["light", "dark"] as const) {
    const colors = (value as ThemePalette)[mode]
    if (
      !colors ||
      Object.keys(colors).length !== Object.keys(defaults[mode]).length
    )
      throw new Error("Invalid theme")
    for (const token of Object.keys(defaults[mode]) as ThemeToken[]) {
      if (
        typeof colors[token] !== "string" ||
        !colorPattern.test(colors[token])
      )
        throw new Error("Invalid theme")
    }
  }
  return value as ThemePalette
}

export function buildThemeCSS(value: ThemePalette) {
  const palette = parseTheme(value)
  return (["light", "dark"] as const)
    .map(
      (mode) =>
        `${mode === "light" ? ":root" : ".dark"}{${Object.entries(palette[mode])
          .map(([token, color]) => `--${token}:${color};`)
          .join("")}}`
    )
    .join("\n")
}

export function applyTheme(palette: ThemePalette) {
  const css = buildThemeCSS(palette)
  let style = document.getElementById("site-theme")
  if (!style) {
    style = document.createElement("style")
    style.id = "site-theme"
    document.head.append(style)
  }
  style.textContent = css
}

export async function loadSiteTheme() {
  let palette = defaultTheme
  try {
    const response = await fetch("/api/site-theme", {
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    })
    if (response.ok) palette = parseTheme(await response.json())
  } catch {
    // Keep the built-in palette while the API is unavailable.
  }
  applyTheme(palette)
}

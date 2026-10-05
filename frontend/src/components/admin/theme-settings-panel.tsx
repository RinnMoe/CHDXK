import { copy } from "@/lib/copy"

import { useState, type CSSProperties } from "react"
import {
  useAdminSystemSettings,
  useUpdateSystemSetting,
} from "@/hooks/use-system-settings"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  applyTheme,
  defaultTheme,
  parseTheme,
  themeSettingKey,
  type ThemePalette,
  type ThemeMode,
  type ThemeToken,
} from "@/lib/site-theme"
import { getErrorMessage } from "./admin-utils"

const common: [ThemeToken, string][] = [
  [
    "primary",
    copy("copy.admin.components.admin.theme-settings-panel.0f132a452ea8"),
  ],
  [
    "primary-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.e3bd6d2293ab"),
  ],
  [
    "background",
    copy("copy.admin.components.admin.theme-settings-panel.c3c6107cf5f5"),
  ],
  [
    "foreground",
    copy("copy.admin.components.admin.theme-settings-panel.09d5cf867e95"),
  ],
  [
    "card",
    copy("copy.admin.components.admin.theme-settings-panel.7a284abb5757"),
  ],
  [
    "card-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.1f160e7c4533"),
  ],
  [
    "muted",
    copy("copy.admin.components.admin.theme-settings-panel.73769529b75c"),
  ],
  [
    "muted-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.1ac73eddc8e4"),
  ],
  [
    "border",
    copy("copy.admin.components.admin.theme-settings-panel.005378329829"),
  ],
  [
    "rating",
    copy("copy.admin.components.admin.theme-settings-panel.f20a2f20a68e"),
  ],
  [
    "success",
    copy("copy.admin.components.admin.theme-settings-panel.bd14e29d1bfe"),
  ],
  [
    "warning",
    copy("copy.admin.components.admin.theme-settings-panel.76e1cfb2b121"),
  ],
  [
    "destructive",
    copy("copy.admin.components.admin.theme-settings-panel.e30a1690a501"),
  ],
  [
    "info",
    copy("copy.admin.components.admin.theme-settings-panel.7c8eb931dcf2"),
  ],
]
const advanced: [ThemeToken, string][] = [
  [
    "popover",
    copy("copy.admin.components.admin.theme-settings-panel.4429849041fe"),
  ],
  [
    "popover-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.24d11ea465b9"),
  ],
  [
    "secondary",
    copy("copy.admin.components.admin.theme-settings-panel.c4ae7ae199e3"),
  ],
  [
    "secondary-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.f6ed56c72add"),
  ],
  [
    "accent",
    copy("copy.admin.components.admin.theme-settings-panel.f8e71ab4e3d1"),
  ],
  [
    "accent-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.594226193acc"),
  ],
  [
    "input",
    copy("copy.admin.components.admin.theme-settings-panel.37a0b2e05e41"),
  ],
  [
    "ring",
    copy("copy.admin.components.admin.theme-settings-panel.0adba3e0ab07"),
  ],
  [
    "chart-1",
    copy("copy.admin.components.admin.theme-settings-panel.17371f0212de"),
  ],
  [
    "chart-2",
    copy("copy.admin.components.admin.theme-settings-panel.d486e988b131"),
  ],
  [
    "chart-3",
    copy("copy.admin.components.admin.theme-settings-panel.d66ba27b63ad"),
  ],
  [
    "chart-4",
    copy("copy.admin.components.admin.theme-settings-panel.a657a95bbb1f"),
  ],
  [
    "chart-5",
    copy("copy.admin.components.admin.theme-settings-panel.790b73aa716b"),
  ],
  [
    "sidebar",
    copy("copy.admin.components.admin.theme-settings-panel.4f8e587e4095"),
  ],
  [
    "sidebar-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.2c2ba43db7a7"),
  ],
  [
    "sidebar-primary",
    copy("copy.admin.components.admin.theme-settings-panel.f8a406d08281"),
  ],
  [
    "sidebar-primary-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.fe1279e4828b"),
  ],
  [
    "sidebar-accent",
    copy("copy.admin.components.admin.theme-settings-panel.c5f6b422fa54"),
  ],
  [
    "sidebar-accent-foreground",
    copy("copy.admin.components.admin.theme-settings-panel.5549e2f7fc67"),
  ],
  [
    "sidebar-border",
    copy("copy.admin.components.admin.theme-settings-panel.15a1c1c3a840"),
  ],
  [
    "sidebar-ring",
    copy("copy.admin.components.admin.theme-settings-panel.67757fa12472"),
  ],
]

function ThemeEditor({ saved }: { saved: string }) {
  const [palette, setPalette] = useState<ThemePalette>(() =>
    parseTheme(JSON.parse(saved))
  )
  const [mode, setMode] = useState<ThemeMode>("light")
  const [error, setError] = useState("")
  const mutation = useUpdateSystemSetting()
  const colors = palette[mode]
  const previewStyle = Object.fromEntries(
    Object.entries(colors).map(([key, value]) => [`--${key}`, value])
  ) as CSSProperties
  const dirty =
    JSON.stringify(palette) !== JSON.stringify(parseTheme(JSON.parse(saved)))
  function fields(items: [ThemeToken, string][]) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {items.map(([token, label]) => (
          <div key={token} className="space-y-2">
            <Label htmlFor={`theme-${mode}-${token}`}>{label}</Label>
            <div className="flex gap-2">
              <Input
                type="color"
                aria-label={label}
                value={colors[token].slice(0, 7)}
                onChange={(event) =>
                  setPalette({
                    ...palette,
                    [mode]: { ...colors, [token]: event.target.value },
                  })
                }
                className="h-10 w-12 shrink-0 p-1"
              />
              <Input
                id={`theme-${mode}-${token}`}
                aria-label={copy(
                  "copy.admin.components.admin.theme-settings-panel.994707c28948",
                  { v0: label }
                )}
                value={colors[token]}
                maxLength={9}
                onChange={(event) =>
                  setPalette({
                    ...palette,
                    [mode]: { ...colors, [token]: event.target.value },
                  })
                }
              />
            </div>
          </div>
        ))}
      </div>
    )
  }
  async function save() {
    setError("")
    try {
      const valid = parseTheme(palette)
      await mutation.mutateAsync({
        key: themeSettingKey,
        cmd: { value: JSON.stringify(valid) },
      })
      applyTheme(valid)
    } catch (err) {
      setError(
        err instanceof Error && err.message === "Invalid theme"
          ? copy(
              "copy.admin.components.admin.theme-settings-panel.ec2d9b307796"
            )
          : getErrorMessage(err)
      )
    }
  }
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap gap-2">
        {(["light", "dark"] as const).map((value) => (
          <Button
            key={value}
            variant={value === mode ? "default" : "outline"}
            onClick={() => setMode(value)}
          >
            {value === "light"
              ? copy(
                  "copy.admin.components.admin.theme-settings-panel.d31a4d6bad47"
                )
              : copy(
                  "copy.admin.components.admin.theme-settings-panel.94d2bd014dfd"
                )}
          </Button>
        ))}
      </div>
      <div
        style={previewStyle}
        className="space-y-3 rounded-lg border bg-background p-5 text-foreground"
      >
        <h4 className="font-semibold">
          {copy(
            "copy.admin.components.admin.theme-settings-panel.b1a5bddaccdc"
          )}
        </h4>
        <p className="text-muted-foreground">
          {copy(
            "copy.admin.components.admin.theme-settings-panel.17b8e8e70b07"
          )}
        </p>
        <div className="rounded-md border bg-card p-4 text-card-foreground">
          <p>
            {copy(
              "copy.admin.components.admin.theme-settings-panel.a5c827d1259d"
            )}
          </p>
          <p className="text-rating">★★★★★</p>
          <div className="mt-2 flex flex-wrap gap-3">
            <span className="text-success">
              {copy(
                "copy.admin.components.admin.theme-settings-panel.cda079bcd55d"
              )}
            </span>
            <span className="text-warning">
              {copy(
                "copy.admin.components.admin.theme-settings-panel.d6e3b43107c5"
              )}
            </span>
            <span className="text-destructive">
              {copy(
                "copy.admin.components.admin.theme-settings-panel.c91c656aae2b"
              )}
            </span>
          </div>
        </div>
        <Button>
          {copy(
            "copy.admin.components.admin.theme-settings-panel.c38a08ef511c"
          )}
        </Button>
      </div>
      {fields(common)}
      <details className="rounded-md border p-4">
        <summary className="cursor-pointer font-medium">
          {copy(
            "copy.admin.components.admin.theme-settings-panel.c71ecb658ee6"
          )}
        </summary>
        <div className="mt-4">{fields(advanced)}</div>
      </details>
      <p className="text-sm text-muted-foreground">
        {copy("copy.admin.components.admin.theme-settings-panel.a152b7f559f6")}
      </p>
      <div className="flex gap-2">
        <Button
          disabled={!dirty || mutation.isPending}
          onClick={() => void save()}
        >
          {mutation.isPending
            ? copy(
                "copy.admin.components.admin.theme-settings-panel.a1a1720d3eb5"
              )
            : copy(
                "copy.admin.components.admin.theme-settings-panel.6997a520f014"
              )}
        </Button>
        <Button
          variant="outline"
          disabled={mutation.isPending}
          onClick={() => setPalette(structuredClone(defaultTheme))}
        >
          {copy(
            "copy.admin.components.admin.theme-settings-panel.6a0ae6668795"
          )}
        </Button>
      </div>
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {copy(
            "copy.admin.components.admin.theme-settings-panel.da6de668e59d"
          )}
          {error}
        </p>
      )}
    </div>
  )
}

export function ThemeSettingsPanel() {
  const { data, isLoading, isError } = useAdminSystemSettings()
  const setting = data?.find((item) => item.key === themeSettingKey)
  return (
    <section className="space-y-4">
      <h3 className="text-lg font-medium">
        {copy("copy.admin.components.admin.theme-settings-panel.6b6b25f98162")}
      </h3>
      <p className="text-sm text-muted-foreground">
        {copy("copy.admin.components.admin.theme-settings-panel.578dbff684af")}
      </p>
      {isLoading ? (
        <p>
          {copy(
            "copy.admin.components.admin.theme-settings-panel.d04fcbda737f"
          )}
        </p>
      ) : isError || !setting ? (
        <p role="alert">
          {copy(
            "copy.admin.components.admin.theme-settings-panel.688fe7c1b333"
          )}
        </p>
      ) : (
        <ThemeEditor key={setting.value} saved={setting.value} />
      )}
    </section>
  )
}

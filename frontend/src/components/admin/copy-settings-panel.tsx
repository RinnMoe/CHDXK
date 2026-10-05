import { copy } from "@/lib/copy"
import { useState } from "react"
import {
  useAdminSystemSettings,
  useUpdateSystemSetting,
} from "@/hooks/use-system-settings"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { getErrorMessage } from "./admin-utils"
import type { SystemSettingDTO } from "@/api/system-settings"

function CopyEditor({ setting }: { setting: SystemSettingDTO }) {
  const [draft, setDraft] = useState<string | null>(null)
  const [error, setError] = useState("")
  const mutation = useUpdateSystemSetting()
  const value = draft ?? setting.value
  async function save() {
    setError("")
    try {
      await mutation.mutateAsync({ key: setting.key, cmd: { value } })
      setDraft(null)
      if (setting.key === "copy.user.config.brand.743e9916a4b2")
        window.location.reload()
    } catch (err) {
      setError(getErrorMessage(err))
    }
  }
  return (
    <article className="space-y-3 rounded-md border p-4">
      <Label htmlFor={`copy-${setting.key}`}>
        {setting.label ||
          copy("copy.admin.components.admin.copy-settings-panel.215bee80e894")}
      </Label>
      <Textarea
        id={`copy-${setting.key}`}
        value={value}
        onChange={(event) => setDraft(event.target.value)}
        disabled={mutation.isPending}
      />
      <div className="flex gap-2">
        <Button
          size="sm"
          onClick={() => void save()}
          disabled={mutation.isPending || value === setting.value}
        >
          {mutation.isPending
            ? copy(
                "copy.admin.components.admin.copy-settings-panel.a1a1720d3eb5"
              )
            : copy(
                "copy.admin.components.admin.copy-settings-panel.8319f71b1df4"
              )}
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={() => setDraft(setting.default_value)}
          disabled={mutation.isPending || value === setting.default_value}
        >
          {copy("copy.admin.components.admin.copy-settings-panel.f242c947972c")}
        </Button>
      </div>
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
    </article>
  )
}

export function CopySettingsPanel() {
  const { data, isLoading, isError } = useAdminSystemSettings()
  const [search, setSearch] = useState("")
  const settings = (data ?? []).filter(
    (setting) =>
      setting.key.startsWith("copy.") &&
      `${setting.label} ${setting.value} ${setting.default_value}`
        .toLocaleLowerCase()
        .includes(search.toLocaleLowerCase())
  )
  return (
    <section className="space-y-4">
      <h3 className="text-lg font-medium">
        {copy("copy.admin.components.admin.copy-settings-panel.12abdae1e423")}
      </h3>
      <p className="text-sm text-muted-foreground">
        {copy("copy.admin.components.admin.copy-settings-panel.6fe1f12e3174")}
      </p>
      <div className="flex flex-wrap items-center gap-3">
        <Input
          aria-label={copy(
            "copy.admin.components.admin.copy-settings-panel.09c2f90e0412"
          )}
          placeholder={copy(
            "copy.admin.components.admin.copy-settings-panel.ebbabcf0b4e4"
          )}
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
          }}
          className="max-w-sm"
        />
      </div>
      {isLoading ? (
        <p>
          {copy("copy.admin.components.admin.copy-settings-panel.ddf1d1f75c40")}
        </p>
      ) : isError ? (
        <p role="alert">
          {copy("copy.admin.components.admin.copy-settings-panel.05e9d77d5fea")}
        </p>
      ) : settings.length === 0 ? (
        <p>
          {copy("copy.admin.components.admin.copy-settings-panel.9fd8c1e32d70")}
        </p>
      ) : (
        <div className="space-y-3">
          {copyGroups.map((group) => {
            const items = settings
              .filter(group.matches)
              .sort((a, b) =>
                a.key.includes("config.brand")
                  ? -1
                  : b.key.includes("config.brand")
                    ? 1
                    : a.label.localeCompare(b.label)
              )
            return items.length ? (
              <section key={group.title} className="space-y-3">
                <h3 className="font-medium">{group.title}</h3>
                <p className="text-sm text-muted-foreground">{group.impact}</p>
                {items.map((setting) => (
                  <CopyEditor key={setting.key} setting={setting} />
                ))}
              </section>
            ) : null
          })}
        </div>
      )}
    </section>
  )
}

const copyGroups = [
  {
    title: copy("copy.admin.components.admin.copy-settings-panel.92ad14acbea2"),
    impact: copy(
      "copy.admin.components.admin.copy-settings-panel.4203beaa87a6"
    ),
    matches: (item: SystemSettingDTO) =>
      item.key.includes("config.brand") ||
      item.key.startsWith("copy.brand.") ||
      item.key.startsWith("copy.pwa."),
  },
  {
    title: copy("copy.admin.components.admin.copy-settings-panel.94ffe9687e66"),
    impact: copy(
      "copy.admin.components.admin.copy-settings-panel.44b1d8d102b1"
    ),
    matches: (item: SystemSettingDTO) => item.key.includes("auth.login-form"),
  },
  {
    title: copy("copy.admin.components.admin.copy-settings-panel.ce7ce316bb6a"),
    impact: copy(
      "copy.admin.components.admin.copy-settings-panel.d59765ed6626"
    ),
    matches: (item: SystemSettingDTO) => item.key.includes("hot-courses-page"),
  },
  {
    title: copy("copy.admin.components.admin.copy-settings-panel.6b0fa56f4d76"),
    impact: copy(
      "copy.admin.components.admin.copy-settings-panel.f444cbeba761"
    ),
    matches: (item: SystemSettingDTO) => item.key.startsWith("copy.email."),
  },
]

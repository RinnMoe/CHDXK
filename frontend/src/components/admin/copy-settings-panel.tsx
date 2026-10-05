import { copy, copyDirectory } from "@/lib/copy"
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

const audienceLabels: Record<string, string> = {
  user: copy("copy.admin.components.admin.copy-settings-panel.f6a2faaac200"),
  admin: copy("copy.admin.components.admin.copy-settings-panel.35748266b6c2"),
  technical: copy(
    "copy.admin.components.admin.copy-settings-panel.3cf925810572"
  ),
}
const pageSize = 20

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
      <p className="text-sm text-muted-foreground">{setting.description}</p>
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
      <details className="text-xs text-muted-foreground">
        <summary className="cursor-pointer break-all">{setting.key}</summary>
        <p className="mt-2 break-all">{copyDirectory[setting.key]?.location}</p>
      </details>
    </article>
  )
}

export function CopySettingsPanel() {
  const { data, isLoading, isError } = useAdminSystemSettings()
  const [search, setSearch] = useState("")
  const [audience, setAudience] = useState("")
  const [page, setPage] = useState(0)
  const settings = (data ?? []).filter(
    (setting) =>
      setting.key.startsWith("copy.") &&
      (!audience || setting.group === `copy_${audience}`) &&
      `${setting.label} ${setting.value} ${setting.default_value}`
        .toLocaleLowerCase()
        .includes(search.toLocaleLowerCase())
  )
  const pages = Math.max(1, Math.ceil(settings.length / pageSize))
  const current = Math.min(page, pages - 1)
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
            setPage(0)
          }}
          className="max-w-sm"
        />
        <select
          aria-label={copy(
            "copy.admin.components.admin.copy-settings-panel.3eb1bd9d8e67"
          )}
          value={audience}
          onChange={(event) => {
            setAudience(event.target.value)
            setPage(0)
          }}
          className="h-9 rounded-md border bg-background px-3 text-sm"
        >
          <option value="">
            {copy(
              "copy.admin.components.admin.copy-settings-panel.eb8666303664"
            )}
          </option>
          {Object.entries(audienceLabels).map(([key, label]) => (
            <option key={key} value={key}>
              {label}
            </option>
          ))}
        </select>
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
          {settings
            .slice(current * pageSize, (current + 1) * pageSize)
            .map((setting) => (
              <CopyEditor key={setting.key} setting={setting} />
            ))}
        </div>
      )}
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          disabled={current === 0}
          onClick={() => setPage(current - 1)}
        >
          {copy("copy.admin.components.admin.copy-settings-panel.c9b9ae7a6144")}
        </Button>
        <span className="text-sm text-muted-foreground">
          {copy("copy.admin.components.admin.copy-settings-panel.9552fe3bd366")}
          {current + 1} / {pages}
          {copy("copy.admin.components.admin.copy-settings-panel.f5080bf79e10")}
          {settings.length}
          {copy("copy.admin.components.admin.copy-settings-panel.f9d529eacdf3")}
        </span>
        <Button
          variant="outline"
          size="sm"
          disabled={current >= pages - 1}
          onClick={() => setPage(current + 1)}
        >
          {copy("copy.admin.components.admin.copy-settings-panel.8a8542f69648")}
        </Button>
      </div>
    </section>
  )
}

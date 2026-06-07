import { useMemo, useState } from "react"
import { RiSaveLine } from "@remixicon/react"
import {
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
} from "@/api/system-settings"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useCourseFilters } from "@/hooks/use-course"
import {
  useAdminSystemSettings,
  useUpdateSystemSetting,
} from "@/hooks/use-system-settings"
import { getErrorMessage } from "./admin-utils"

type DraftMap = Record<string, string>

const GROUP_LABELS: Record<string, string> = {
  admin: "管理",
  course: "课程",
  review: "点评",
}

export function SystemSettingsTab() {
  const settingsQuery = useAdminSystemSettings()
  const filtersQuery = useCourseFilters()
  const updateMutation = useUpdateSystemSetting()
  const [drafts, setDrafts] = useState<DraftMap>({})
  const [message, setMessage] = useState("")
  const [error, setError] = useState("")

  const semesters = useMemo(
    () => filtersQuery.data?.semesters?.filter((item) => item.name) ?? [],
    [filtersQuery.data?.semesters]
  )

  const groups = useMemo(() => {
    const items = settingsQuery.data ?? []
    return items.reduce<Record<string, SystemSettingDTO[]>>((acc, item) => {
      const group = item.group || "other"
      acc[group] = [...(acc[group] ?? []), item]
      return acc
    }, {})
  }, [settingsQuery.data])

  function draftValue(setting: SystemSettingDTO) {
    return drafts[setting.key] ?? setting.value
  }

  function setDraft(key: string, value: string) {
    setDrafts((current) => ({ ...current, [key]: value }))
    setMessage("")
    setError("")
  }

  async function saveSetting(setting: SystemSettingDTO) {
    setMessage("")
    setError("")
    try {
      await updateMutation.mutateAsync({
        key: setting.key,
        cmd: { value: draftValue(setting) },
      })
      setDrafts((current) => {
        const next = { ...current }
        delete next[setting.key]
        return next
      })
      setMessage("已保存")
    } catch (err) {
      setError(getErrorMessage(err))
    }
  }

  const isLoading = settingsQuery.isLoading || filtersQuery.isLoading

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-medium">系统设置</h2>

      {error && (
        <p className="text-sm text-destructive" role="alert">
          {error}
        </p>
      )}
      {message && <p className="text-sm text-muted-foreground">{message}</p>}

      <div className="space-y-6">
        {Object.entries(groups).map(([group, settings]) => (
          <section key={group} className="space-y-3">
            <h3 className="text-sm font-medium text-muted-foreground">
              {GROUP_LABELS[group] ?? group}
            </h3>
            <div className="divide-y rounded-md border">
              {settings.map((setting) => {
                const value = draftValue(setting)
                const isDirty = value !== setting.value
                const submitting =
                  updateMutation.isPending &&
                  updateMutation.variables?.key === setting.key

                return (
                  <div
                    key={setting.key}
                    className="grid gap-3 p-4 md:grid-cols-[minmax(0,1fr)_minmax(240px,320px)_auto] md:items-center"
                  >
                    <div className="min-w-0 space-y-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <Label htmlFor={`setting-${setting.key}`}>
                          {setting.label || setting.key}
                        </Label>
                        {setting.public && <Badge variant="outline">公开</Badge>}
                        {setting.requires_restart && (
                          <Badge variant="secondary">需重启</Badge>
                        )}
                      </div>
                      {setting.description && (
                        <p className="text-sm text-muted-foreground">
                          {setting.description}
                        </p>
                      )}
                      <p className="truncate text-xs text-muted-foreground">
                        {setting.key}
                      </p>
                    </div>

                    <SettingControl
                      setting={setting}
                      value={value}
                      semesters={semesters.map((item) => item.name)}
                      disabled={isLoading || submitting}
                      onChange={(next) => setDraft(setting.key, next)}
                    />

                    <Button
                      type="button"
                      size="sm"
                      disabled={isLoading || submitting || !isDirty}
                      onClick={() => void saveSetting(setting)}
                    >
                      <RiSaveLine data-icon="inline-start" />
                      {submitting ? "保存中" : "保存"}
                    </Button>
                  </div>
                )
              })}
            </div>
          </section>
        ))}
      </div>
    </section>
  )
}

function SettingControl({
  setting,
  value,
  semesters,
  disabled,
  onChange,
}: {
  setting: SystemSettingDTO
  value: string
  semesters: string[]
  disabled: boolean
  onChange: (value: string) => void
}) {
  if (setting.key === SYSTEM_SETTING_CURRENT_SEMESTER) {
    return (
      <Select value={value} onValueChange={onChange} disabled={disabled}>
        <SelectTrigger id={`setting-${setting.key}`} className="w-full">
          <SelectValue placeholder="选择学期" />
        </SelectTrigger>
        <SelectContent>
          {semesters.map((semester) => (
            <SelectItem key={semester} value={semester}>
              {semester}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )
  }

  if (setting.type === "bool") {
    return (
      <div className="flex h-9 items-center gap-2">
        <Checkbox
          id={`setting-${setting.key}`}
          checked={value === "true"}
          disabled={disabled}
          onCheckedChange={(checked) =>
            onChange(checked === true ? "true" : "false")
          }
        />
        <span className="text-sm text-muted-foreground">
          启用
        </span>
      </div>
    )
  }

  return (
    <Input
      id={`setting-${setting.key}`}
      type={setting.type === "int" || setting.type === "float" ? "number" : "text"}
      step={setting.type === "float" ? "0.01" : undefined}
      value={value}
      disabled={disabled}
      onChange={(event) => onChange(event.target.value)}
    />
  )
}

import { useMemo, useState } from "react"
import { RiCloseLine, RiSaveLine } from "@remixicon/react"
import {
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
} from "@/api/system-settings"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
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
type ErrorMap = Record<string, string>

const GROUP_LABELS: Record<string, string> = {
  admin: "管理",
  api_key: "API key",
  auth: "认证",
  course: "课程",
  review: "点评",
}

export function SystemSettingsTab() {
  const settingsQuery = useAdminSystemSettings()
  const filtersQuery = useCourseFilters()
  const updateMutation = useUpdateSystemSetting()
  const [drafts, setDrafts] = useState<DraftMap>({})
  const [errors, setErrors] = useState<ErrorMap>({})

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
    setErrors((current) => {
      const next = { ...current }
      delete next[key]
      return next
    })
  }

  function restoreDefault(setting: SystemSettingDTO) {
    setDraft(setting.key, setting.default_value)
  }

  async function saveSetting(setting: SystemSettingDTO) {
    setErrors((current) => {
      const next = { ...current }
      delete next[setting.key]
      return next
    })
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
    } catch (err) {
      setErrors((current) => ({
        ...current,
        [setting.key]: getErrorMessage(err),
      }))
    }
  }

  const isLoading = settingsQuery.isLoading || filtersQuery.isLoading

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-medium">系统设置</h2>

      <TooltipProvider>
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
                  const settingError = errors[setting.key]

                  return (
                    <div
                      key={setting.key}
                      className="grid gap-3 p-4 md:grid-cols-[minmax(0,1fr)_minmax(240px,320px)_5rem] md:items-center"
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

                      <div className="min-w-0 space-y-1">
                        <SettingControl
                          setting={setting}
                          value={value}
                          semesters={semesters.map((item) => item.name)}
                          disabled={isLoading || submitting}
                          onChange={(next) => setDraft(setting.key, next)}
                        />
                        {settingError && (
                          <p className="text-xs text-destructive" role="alert">
                            {settingError}
                          </p>
                        )}
                      </div>

                      <div className="flex h-8 w-20 items-center justify-end gap-2">
                        {isDirty && (
                          <>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  type="button"
                                  size="icon-sm"
                                  aria-label={submitting ? "保存中" : "保存"}
                                  disabled={isLoading || submitting}
                                  onClick={() => void saveSetting(setting)}
                                >
                                  <RiSaveLine />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                {submitting ? "保存中" : "保存"}
                              </TooltipContent>
                            </Tooltip>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  type="button"
                                  size="icon-sm"
                                  variant="outline"
                                  aria-label="恢复默认值"
                                  disabled={
                                    isLoading ||
                                    submitting ||
                                    value === setting.default_value
                                  }
                                  onClick={() => restoreDefault(setting)}
                                >
                                  <RiCloseLine />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>恢复默认值</TooltipContent>
                            </Tooltip>
                          </>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </section>
          ))}
        </div>
      </TooltipProvider>
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
      <div className="flex h-9 items-center">
        <Switch
          id={`setting-${setting.key}`}
          checked={value === "true"}
          disabled={disabled}
          onCheckedChange={(checked) =>
            onChange(checked === true ? "true" : "false")
          }
        />
      </div>
    )
  }

  if (setting.type === "string_list") {
    return (
      <Textarea
        id={`setting-${setting.key}`}
        className="min-h-20 resize-y"
        value={value}
        disabled={disabled}
        onChange={(event) => onChange(event.target.value)}
      />
    )
  }

  return (
    <Input
      id={`setting-${setting.key}`}
      type={
        setting.type === "int" || setting.type === "float" ? "number" : "text"
      }
      step={setting.type === "float" ? "0.01" : undefined}
      value={value}
      disabled={disabled}
      onChange={(event) => onChange(event.target.value)}
    />
  )
}

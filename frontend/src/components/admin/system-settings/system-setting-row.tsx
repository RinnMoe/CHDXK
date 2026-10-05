import { copy } from "@/lib/copy"
import { RiCloseLine, RiSaveLine } from "@remixicon/react"
import type { SystemSettingDTO } from "@/api/system-settings"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { SettingControl } from "./setting-control"

interface SystemSettingRowProps {
  setting: SystemSettingDTO
  value: string
  semesters: string[]
  isDirty: boolean
  isLoading: boolean
  submitting: boolean
  error?: string
  onChange: (value: string) => void
  onSave: () => void
  onRestoreDefault: () => void
}

export function SystemSettingRow({
  setting,
  value,
  semesters,
  isDirty,
  isLoading,
  submitting,
  error,
  onChange,
  onSave,
  onRestoreDefault,
}: SystemSettingRowProps) {
  return (
    <div className="grid gap-3 p-4 md:grid-cols-[minmax(0,1fr)_minmax(240px,320px)_5rem] md:items-center">
      <div className="min-w-0 space-y-1">
        <div className="flex flex-wrap items-center gap-2">
          <Label htmlFor={`setting-${setting.key}`}>
            {setting.label || setting.key}
          </Label>
          {setting.public && (
            <Badge variant="outline">
              {copy(
                "copy.admin.components.admin.system-settings.system-setting-row.7a73baf0ce9a"
              )}
            </Badge>
          )}
          {setting.requires_restart && (
            <Badge variant="secondary">
              {copy(
                "copy.admin.components.admin.system-settings.system-setting-row.7c4436078ee9"
              )}
            </Badge>
          )}
        </div>
        {setting.description && (
          <p className="text-sm text-muted-foreground">{setting.description}</p>
        )}
        <p className="text-xs break-all text-muted-foreground">{setting.key}</p>
      </div>

      <div className="min-w-0 space-y-1">
        <SettingControl
          setting={setting}
          value={value}
          semesters={semesters}
          disabled={isLoading || submitting}
          onChange={onChange}
        />
        {error && (
          <p className="text-xs text-destructive" role="alert">
            {error}
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
                  aria-label={
                    submitting
                      ? copy(
                          "copy.admin.components.admin.system-settings.system-setting-row.a1a1720d3eb5"
                        )
                      : copy(
                          "copy.admin.components.admin.system-settings.system-setting-row.a3030bf8f16d"
                        )
                  }
                  disabled={isLoading || submitting}
                  onClick={onSave}
                >
                  <RiSaveLine />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {submitting
                  ? copy(
                      "copy.admin.components.admin.system-settings.system-setting-row.a1a1720d3eb5"
                    )
                  : copy(
                      "copy.admin.components.admin.system-settings.system-setting-row.a3030bf8f16d"
                    )}
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  size="icon-sm"
                  variant="outline"
                  aria-label={copy(
                    "copy.admin.components.admin.system-settings.system-setting-row.f242c947972c"
                  )}
                  disabled={
                    isLoading || submitting || value === setting.default_value
                  }
                  onClick={onRestoreDefault}
                >
                  <RiCloseLine />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {copy(
                  "copy.admin.components.admin.system-settings.system-setting-row.f242c947972c"
                )}
              </TooltipContent>
            </Tooltip>
          </>
        )}
      </div>
    </div>
  )
}

import {
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
} from "@/api/system-settings"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"

interface SettingControlProps {
  setting: SystemSettingDTO
  value: string
  semesters: string[]
  disabled: boolean
  onChange: (value: string) => void
}

export function SettingControl({
  setting,
  value,
  semesters,
  disabled,
  onChange,
}: SettingControlProps) {
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

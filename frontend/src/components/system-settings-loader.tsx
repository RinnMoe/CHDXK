import { useSystemSettings } from "@/hooks/use-system-settings"

export function SystemSettingsLoader() {
  useSystemSettings(true)
  return null
}

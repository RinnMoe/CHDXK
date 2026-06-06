import { useAuth } from "@/contexts/auth-context"
import { useSystemSettings } from "@/hooks/use-system-settings"

export function SystemSettingsLoader() {
  const { user } = useAuth()
  useSystemSettings(!!user)
  return null
}

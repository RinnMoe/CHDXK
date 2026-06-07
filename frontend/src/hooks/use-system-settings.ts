import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  listAdminSystemSettings,
  listSystemSettings,
  updateSystemSetting,
  SYSTEM_SETTING_AUTH_EMAIL_DOMAIN,
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
  type UpdateSystemSettingCommand,
} from "@/api/system-settings"
import { defaultAuthEmailDomain } from "@/config/auth"

export function useSystemSettings(enabled = true) {
  return useQuery({
    queryKey: ["system-settings", "public"],
    queryFn: listSystemSettings,
    enabled,
  })
}

export function useAdminSystemSettings(enabled = true) {
  return useQuery({
    queryKey: ["system-settings", "admin"],
    queryFn: listAdminSystemSettings,
    enabled,
  })
}

export function useUpdateSystemSetting() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      key,
      cmd,
    }: {
      key: string
      cmd: UpdateSystemSettingCommand
    }) => updateSystemSetting(key, cmd),
    onSuccess: (updated: SystemSettingDTO) => {
      for (const scope of ["public", "admin"] as const) {
        queryClient.setQueryData<SystemSettingDTO[]>(
          ["system-settings", scope],
          (settings) => {
            const current = settings ?? []
            const index = current.findIndex((item) => item.key === updated.key)
            if (index < 0) return updated.public ? [...current, updated] : current
            return current.map((item, i) => (i === index ? updated : item))
          }
        )
      }
      void queryClient.invalidateQueries({ queryKey: ["system-settings"] })
    },
  })
}

export function getCurrentSemesterSetting(
  settings?: SystemSettingDTO[] | null
) {
  return (
    settings?.find((item) => item.key === SYSTEM_SETTING_CURRENT_SEMESTER)
      ?.value ?? ""
  )
}

export function getAuthEmailDomainSetting(
  settings?: SystemSettingDTO[] | null
) {
  return (
    settings?.find((item) => item.key === SYSTEM_SETTING_AUTH_EMAIL_DOMAIN)
      ?.value || defaultAuthEmailDomain
  )
}

export function useAuthEmailDomain() {
  const settingsQuery = useSystemSettings()
  return getAuthEmailDomainSetting(settingsQuery.data)
}

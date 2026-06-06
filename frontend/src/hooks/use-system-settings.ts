import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  listSystemSettings,
  updateSystemSetting,
  SYSTEM_SETTING_CURRENT_SEMESTER,
  type SystemSettingDTO,
  type UpdateSystemSettingCommand,
} from "@/api/system-settings"

export function useSystemSettings(enabled = true) {
  return useQuery({
    queryKey: ["system-settings"],
    queryFn: listSystemSettings,
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
      queryClient.setQueryData<SystemSettingDTO[]>(
        ["system-settings"],
        (settings) => {
          const current = settings ?? []
          const index = current.findIndex((item) => item.key === updated.key)
          if (index < 0) return [...current, updated]
          return current.map((item, i) => (i === index ? updated : item))
        }
      )
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

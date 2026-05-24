import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  getUserSettings,
  updateUserSettings,
  type UpdateUserSettingsCommand,
  type UserSettingsDTO,
} from "@/api/user-settings"

export function useUserSettings(enabled = true) {
  return useQuery({
    queryKey: ["user-settings"],
    queryFn: getUserSettings,
    enabled,
  })
}

export function useUpdateUserSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (cmd: UpdateUserSettingsCommand) => updateUserSettings(cmd),
    onSuccess: (settings: UserSettingsDTO) => {
      queryClient.setQueryData(["user-settings"], settings)
    },
  })
}

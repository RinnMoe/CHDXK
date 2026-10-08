import { useSystemSettings } from "./use-system-settings"

export function useImportSettings() {
  const query = useSystemSettings()
  const enabled = (key: string) =>
    !query.isPending &&
    !query.isError &&
    (query.data?.find((item) => item.key === key)?.value ?? "true") === "true"
  return {
    school: enabled("import.school.enabled"),
    manual: enabled("import.manual.enabled"),
    mychu: enabled("import.mychu.visible"),
    admin: enabled("import.admin.enabled"),
  }
}

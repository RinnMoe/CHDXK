import { useMemo, useState } from "react"
import { RiSaveLine } from "@remixicon/react"
import { SYSTEM_SETTING_CURRENT_SEMESTER } from "@/api/system-settings"
import { Button } from "@/components/ui/button"
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
  getCurrentSemesterSetting,
  useSystemSettings,
  useUpdateSystemSetting,
} from "@/hooks/use-system-settings"
import { getErrorMessage } from "./admin-utils"

export function SystemSettingsTab() {
  const settingsQuery = useSystemSettings()
  const filtersQuery = useCourseFilters()
  const updateMutation = useUpdateSystemSetting()
  const savedSemester = getCurrentSemesterSetting(settingsQuery.data)
  const [draftSemester, setDraftSemester] = useState<string | null>(null)
  const [message, setMessage] = useState("")
  const [error, setError] = useState("")

  const semesters = useMemo(
    () => filtersQuery.data?.semesters?.filter((item) => item.name) ?? [],
    [filtersQuery.data?.semesters]
  )

  const currentSemester = draftSemester ?? savedSemester

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setMessage("")
    setError("")
    try {
      await updateMutation.mutateAsync({
        key: SYSTEM_SETTING_CURRENT_SEMESTER,
        cmd: { value: currentSemester },
      })
      setDraftSemester(null)
      setMessage("已保存")
    } catch (err) {
      setError(getErrorMessage(err))
    }
  }

  const isLoading = settingsQuery.isLoading || filtersQuery.isLoading
  const isDirty = draftSemester !== null && currentSemester !== savedSemester
  const submitting = updateMutation.isPending

  return (
    <section className="space-y-4">
      <h2 className="text-lg font-medium">系统设置</h2>

      <form className="max-w-md space-y-4" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <Label htmlFor="system-current-semester">当前学期</Label>
          <Select
            value={currentSemester}
            onValueChange={(value) => {
              setDraftSemester(value)
              setMessage("")
              setError("")
            }}
            disabled={isLoading || semesters.length === 0}
          >
            <SelectTrigger id="system-current-semester" className="w-full">
              <SelectValue placeholder="选择学期" />
            </SelectTrigger>
            <SelectContent>
              {semesters.map((semester) => (
                <SelectItem key={semester.name} value={semester.name}>
                  {semester.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {error && (
          <p className="text-sm text-destructive" role="alert">
            {error}
          </p>
        )}
        {message && <p className="text-sm text-muted-foreground">{message}</p>}

        <Button
          type="submit"
          disabled={isLoading || submitting || !isDirty || !currentSemester}
        >
          <RiSaveLine data-icon="inline-start" />
          {submitting ? "保存中" : "保存"}
        </Button>
      </form>
    </section>
  )
}

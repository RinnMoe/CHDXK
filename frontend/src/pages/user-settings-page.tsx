import { useMemo, useState } from "react"
import { useForm } from "@tanstack/react-form"
import { RiSaveLine } from "@remixicon/react"
import { fieldError } from "@/components/auth/form-utils"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/contexts/auth-context"
import { useCourseFilters } from "@/hooks/use-course"
import {
  useUpdateUserSettings,
  useUserSettings,
} from "@/hooks/use-user-settings"
import { useTheme } from "@/components/theme-provider"

const themeOptions = [
  { value: "system", label: "跟随系统" },
  { value: "light", label: "亮色" },
  { value: "dark", label: "暗色" },
] as const

type UserSettingsFormValues = {
  currentSemester: string
}

type SemesterOption = {
  name: string
}

function ThemeSettings() {
  const { theme, setTheme } = useTheme()

  function handleThemeChange(value: string) {
    if (value === "system" || value === "light" || value === "dark") {
      setTheme(value)
    }
  }

  return (
    <div className="space-y-2">
      <Label htmlFor="theme-mode">外观模式</Label>
      <Select value={theme} onValueChange={handleThemeChange}>
        <SelectTrigger id="theme-mode" className="w-full">
          <SelectValue placeholder="选择外观模式" />
        </SelectTrigger>
        <SelectContent>
          {themeOptions.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <p className="text-sm leading-6 text-muted-foreground">
        仅保存在当前浏览器，不会同步到其他设备。
      </p>
    </div>
  )
}

function UserSettingsForm({
  initialSemester,
  semesters,
}: {
  initialSemester: string
  semesters: SemesterOption[]
}) {
  const updateMutation = useUpdateUserSettings()
  const [message, setMessage] = useState("")
  const [submitError, setSubmitError] = useState("")

  const form = useForm({
    defaultValues: {
      currentSemester: initialSemester,
    } as UserSettingsFormValues,
    onSubmit: async ({ value }) => {
      const settings = await updateMutation.mutateAsync({
        current_semester: value.currentSemester,
      })
      setMessage("已保存")
      return settings
    },
  })

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    e.stopPropagation()
    setMessage("")
    setSubmitError("")
    void form.handleSubmit().catch((err: unknown) => {
      setSubmitError(err instanceof Error ? err.message : "保存失败")
    })
  }

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <form.Field
        name="currentSemester"
        validators={{
          onSubmit: ({ value }) => (value ? undefined : "请选择当前学期"),
        }}
      >
        {(field) => {
          const error = fieldError(field.state.meta.errors)
          return (
            <div className="space-y-2">
              <Label htmlFor="current-semester">当前学期</Label>
              <Select
                value={field.state.value}
                onValueChange={(value) => {
                  field.handleChange(value)
                  setMessage("")
                  setSubmitError("")
                }}
                disabled={semesters.length === 0}
              >
                <SelectTrigger id="current-semester" className="w-full">
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
              <p className="text-sm leading-6 text-muted-foreground">
                影响课程相关页面的默认学期选择。
              </p>
              {error && (
                <p className="text-sm text-destructive" role="alert">
                  {error}
                </p>
              )}
            </div>
          )
        }}
      </form.Field>

      {submitError && (
        <p className="text-sm text-destructive" role="alert">
          {submitError}
        </p>
      )}
      {message && <p className="text-sm text-muted-foreground">{message}</p>}

      <div className="flex justify-end">
        <form.Subscribe
          selector={(state) => ({
            isSubmitting: state.isSubmitting,
            currentSemester: state.values.currentSemester,
          })}
        >
          {({ isSubmitting, currentSemester }) => {
            const isDirty = currentSemester !== initialSemester
            const submitting = isSubmitting || updateMutation.isPending
            return (
              <Button
                type="submit"
                disabled={submitting || !isDirty || semesters.length === 0}
              >
                <RiSaveLine data-icon="inline-start" />
                {submitting ? "保存中" : "保存"}
              </Button>
            )
          }}
        </form.Subscribe>
      </div>
    </form>
  )
}

export function UserSettingsPage() {
  const { user } = useAuth()
  const settingsQuery = useUserSettings(!!user)
  const filtersQuery = useCourseFilters()

  const semesters = useMemo(
    () => filtersQuery.data?.semesters?.filter((item) => item.name) ?? [],
    [filtersQuery.data?.semesters]
  )

  const isLoading = settingsQuery.isLoading || filtersQuery.isLoading
  const savedSemester = settingsQuery.data?.current_semester ?? ""

  return (
    <>
      <PageTitle>用户设置</PageTitle>
      <PageShell>
        <div className="mx-auto max-w-2xl space-y-6">
          <h1 className="text-2xl font-semibold">用户设置</h1>

          <Card>
            <CardHeader>
              <CardTitle>偏好设置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <section className="space-y-3">
                <h2 className="text-sm font-medium">外观</h2>
                <ThemeSettings />
              </section>

              <Separator />

              <section className="space-y-3">
                <h2 className="text-sm font-medium">课程</h2>
                {isLoading ? (
                  <div className="space-y-3">
                    <Skeleton className="h-5 w-20" />
                    <Skeleton className="h-9 w-full" />
                  </div>
                ) : (
                  <UserSettingsForm
                    initialSemester={savedSemester}
                    semesters={semesters}
                  />
                )}
              </section>
            </CardContent>
          </Card>
        </div>
      </PageShell>
    </>
  )
}

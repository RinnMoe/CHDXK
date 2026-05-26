import { useMemo, useState } from "react"
import { Navigate } from "react-router-dom"
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
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/contexts/auth-context"
import { useCourseFilters } from "@/hooks/use-course"
import { useLoginRedirectPath } from "@/hooks/use-login-redirect"
import {
  useUpdateUserSettings,
  useUserSettings,
} from "@/hooks/use-user-settings"

type UserSettingsFormValues = {
  currentSemester: string
}

type SemesterOption = {
  name: string
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
  const { user, isLoading: authLoading } = useAuth()
  const loginRedirectPath = useLoginRedirectPath()
  const settingsQuery = useUserSettings(!!user)
  const filtersQuery = useCourseFilters()

  const semesters = useMemo(
    () => filtersQuery.data?.semesters?.filter((item) => item.name) ?? [],
    [filtersQuery.data?.semesters]
  )

  if (authLoading) return null
  if (!user) return <Navigate to={loginRedirectPath} replace />

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
              <CardTitle>课程</CardTitle>
            </CardHeader>
            <CardContent>
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
            </CardContent>
          </Card>
        </div>
      </PageShell>
    </>
  )
}

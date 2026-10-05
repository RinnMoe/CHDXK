import { useAuth } from "@/contexts/auth-context"
import { getAccountBindingURL } from "@/api/auth"
import { Button } from "@/components/ui/button"
import { copy } from "@/lib/copy"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useTheme } from "@/components/theme-provider"

const themeOptions = [
  {
    value: "system",
    label: copy("copy.user.pages.user-settings-page.217cfe7db1e3"),
  },
  {
    value: "light",
    label: copy("copy.user.pages.user-settings-page.300cb577254e"),
  },
  {
    value: "dark",
    label: copy("copy.user.pages.user-settings-page.29edc1452eff"),
  },
] as const

function ThemeSettings() {
  const { theme, setTheme } = useTheme()

  function handleThemeChange(value: string) {
    if (value === "system" || value === "light" || value === "dark") {
      setTheme(value)
    }
  }

  return (
    <div className="space-y-2">
      <Label htmlFor="theme-mode">
        {copy("copy.user.pages.user-settings-page.97cefbbbfb55")}
      </Label>
      <Select value={theme} onValueChange={handleThemeChange}>
        <SelectTrigger id="theme-mode" className="w-full">
          <SelectValue
            placeholder={copy(
              "copy.user.pages.user-settings-page.bf643e1be362"
            )}
          />
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
        {copy("copy.user.pages.user-settings-page.1e7d7c9e9437")}
      </p>
    </div>
  )
}

export function UserSettingsPage() {
  const { user } = useAuth()
  return (
    <>
      <PageTitle>
        {copy("copy.user.pages.user-settings-page.7ce710f646db")}
      </PageTitle>
      <PageShell>
        <div className="mx-auto max-w-2xl space-y-6">
          <h1 className="text-2xl font-semibold">
            {copy("copy.user.pages.user-settings-page.7ce710f646db")}
          </h1>

          {user?.is_bootstrap && (
            <Card>
              <CardHeader>
                <CardTitle>
                  {copy("copy.user.pages.user-settings-page.273af86c2ffa")}
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  {user.subject
                    ? copy("copy.user.pages.user-settings-page.c7444419a769", {
                        v0: user.bound_nickname || "频道账号",
                      })
                    : copy("copy.user.pages.user-settings-page.60971d3b32b5")}
                </p>
                <p className="text-sm text-muted-foreground">
                  {copy("copy.user.pages.user-settings-page.7bb8c99f8a2e")}
                </p>
                <Button
                  onClick={() => window.location.assign(getAccountBindingURL())}
                >
                  {user.subject
                    ? copy("copy.user.pages.user-settings-page.9a5ba76dcae9")
                    : copy("copy.user.pages.user-settings-page.a89a3f0314b6")}
                </Button>
              </CardContent>
            </Card>
          )}
          <Card>
            <CardHeader>
              <CardTitle>
                {copy("copy.user.pages.user-settings-page.57907a69fc5a")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <section>
                <ThemeSettings />
              </section>
            </CardContent>
          </Card>
        </div>
      </PageShell>
    </>
  )
}

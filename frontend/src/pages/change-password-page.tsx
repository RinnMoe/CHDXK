import { PasswordForm } from "@/components/auth/password-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { useAuth } from "@/contexts/auth-context"
import { Button } from "@/components/ui/button"

export function ChangePasswordPage() {
  const { logout } = useAuth()
  return (
    <>
      <PageTitle>修改初始密码</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="mx-auto max-w-sm space-y-6 py-8">
          <h1 className="text-xl font-semibold">请先修改初始密码</h1>
          <p className="text-sm text-muted-foreground">
            设置新密码后即可继续使用本站。
          </p>
          <PasswordForm required />
          <Button
            variant="ghost"
            onClick={() =>
              void logout().then(() => window.location.replace("/login"))
            }
          >
            退出登录
          </Button>
        </div>
      </PageShell>
    </>
  )
}

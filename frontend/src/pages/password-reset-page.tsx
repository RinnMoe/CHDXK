import { PasswordResetForm } from "@/components/auth/password-reset-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"

export function PasswordResetPage() {
  return (
    <>
      <PageTitle>重置密码</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <PasswordResetForm />
        </div>
      </PageShell>
    </>
  )
}

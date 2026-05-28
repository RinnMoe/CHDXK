import { RegisterForm } from "@/components/auth/register-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"

export function RegisterPage() {
  return (
    <>
      <PageTitle>注册</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <RegisterForm />
        </div>
      </PageShell>
    </>
  )
}

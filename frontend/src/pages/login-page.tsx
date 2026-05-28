import { AboutAgreementExcerpt } from "@/components/about/about-content"
import { LoginForm } from "@/components/auth/login-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"

export function LoginPage() {
  return (
    <>
      <PageTitle>登录</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <LoginForm />
        </div>
        <AboutAgreementExcerpt />
      </PageShell>
    </>
  )
}

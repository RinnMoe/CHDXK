import { copy } from "@/lib/copy"
import { AboutAgreementExcerpt } from "@/components/about/about-content"
import { LoginForm } from "@/components/auth/login-form"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"

export function LoginPage() {
  return (
    <>
      <PageTitle>{copy("copy.user.pages.login-page.1e2df9c3075a")}</PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="flex justify-center">
          <LoginForm />
        </div>
        <AboutAgreementExcerpt />
      </PageShell>
    </>
  )
}

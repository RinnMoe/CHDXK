import { copy } from "@/lib/copy"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { AboutContent } from "@/components/about/about-content"

export function AboutPage() {
  return (
    <>
      <PageTitle>{copy("copy.user.pages.about-page.52d25a9e30ba")}</PageTitle>
      <PageShell>
        <AboutContent />
      </PageShell>
    </>
  )
}

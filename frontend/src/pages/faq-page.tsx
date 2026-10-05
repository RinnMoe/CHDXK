import { copy } from "@/lib/copy"

import { PageShell } from "@/components/layout/page-shell"
import { SitePageContent } from "@/components/common/site-page-content"
import { PageTitle } from "@/components/common/page-title"

export function FaqPage() {
  return (
    <>
      <PageTitle>{copy("copy.user.pages.faq-page.45a6d115fdfb")}</PageTitle>
      <PageShell>
        <SitePageContent slug="faq" />
      </PageShell>
    </>
  )
}

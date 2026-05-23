import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { AboutContent } from "@/components/about/about-content"

export function AboutPage() {
  return (
    <>
      <PageTitle>关于</PageTitle>
      <PageShell>
        <AboutContent />
      </PageShell>
    </>
  )
}

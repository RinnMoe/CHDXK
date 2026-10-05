import { copy } from "@/lib/copy"

import { useParams } from "@tanstack/react-router"
import { useSitePages } from "@/api/site-pages"
import { SitePageContent } from "@/components/common/site-page-content"
import { PageTitle } from "@/components/common/page-title"
import { PageShell } from "@/components/layout/page-shell"

export function SitePage() {
  const { slug } = useParams({ strict: false })
  const pages = useSitePages()
  return (
    <>
      <PageTitle>
        {pages.data?.find((p) => p.slug === slug)?.title ??
          copy("copy.user.pages.site-page.452c7b10d57a")}
      </PageTitle>
      <PageShell>
        <SitePageContent slug={slug ?? ""} />
      </PageShell>
    </>
  )
}

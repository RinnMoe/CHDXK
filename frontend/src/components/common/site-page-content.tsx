import { copy } from "@/lib/copy"

import { useSitePages } from "@/api/site-pages"
import { SafeMarkdown } from "@/components/review/safe-markdown"
import { brand } from "@/config/brand"

export function SitePageContent({ slug }: { slug: string }) {
  const query = useSitePages()
  const page = query.data?.find((item) => item.slug === slug)
  if (query.isLoading)
    return (
      <p>
        {copy("copy.user.components.common.site-page-content.4927a53bcc88")}
      </p>
    )
  if (query.isError)
    return (
      <p role="alert">
        {copy("copy.user.components.common.site-page-content.13300419ebf5")}
      </p>
    )
  if (!page)
    return (
      <p>
        {copy("copy.user.components.common.site-page-content.2edeec8c0777")}
      </p>
    )
  return (
    <article className="prose max-w-none dark:prose-invert">
      <SafeMarkdown
        content={page.content.replaceAll("{site_name}", brand.name)}
      />
    </article>
  )
}

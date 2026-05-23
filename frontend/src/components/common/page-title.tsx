import { formatPageTitle } from "@/config/brand"

type PageTitleProps = {
  children?: string
}

export function PageTitle({ children }: PageTitleProps) {
  return <title>{formatPageTitle(children)}</title>
}

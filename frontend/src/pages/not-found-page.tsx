import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import { Button } from "@/components/ui/button"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"

export function NotFoundPage() {
  return (
    <>
      <PageTitle>
        {copy("copy.user.pages.not-found-page.e75737c1ab90")}
      </PageTitle>
      <PageShell showAnnouncements={false}>
        <div className="mx-auto flex min-h-[calc(100svh-15rem)] max-w-2xl flex-col items-center justify-center py-16 text-center">
          <p className="text-sm font-medium text-primary">404</p>
          <h1 className="mt-3 text-3xl font-bold tracking-normal sm:text-4xl">
            {copy("copy.user.pages.not-found-page.e75737c1ab90")}
          </h1>
          <p className="mt-3 text-sm leading-7 text-muted-foreground">
            {copy("copy.user.pages.not-found-page.c84613ca19d6")}
          </p>
          <div className="mt-6 flex flex-col gap-3 sm:flex-row">
            <Button asChild>
              <Link to="/">
                {copy("copy.user.pages.not-found-page.d66f4359dae3")}
              </Link>
            </Button>
            <Button asChild variant="outline">
              <Link to="/course">
                {copy("copy.user.pages.not-found-page.541d030f3e43")}
              </Link>
            </Button>
          </div>
        </div>
      </PageShell>
    </>
  )
}

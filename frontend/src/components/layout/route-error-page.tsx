import { copy } from "@/lib/copy"
import type { ErrorComponentProps } from "@tanstack/react-router"

import { PageTitle } from "@/components/common/page-title"

export function RouteErrorPage(_props: ErrorComponentProps) {
  return (
    <>
      <PageTitle>
        {copy("copy.user.components.layout.route-error-page.1a079fc9362f")}
      </PageTitle>
      <main className="flex min-h-svh items-center justify-center bg-background px-4 py-12 text-foreground">
        <section className="max-w-xl text-center">
          <div className="inline-flex items-center rounded-full border bg-card px-3 py-1 text-xs font-medium text-muted-foreground shadow-xs">
            {copy("copy.user.components.layout.route-error-page.1a079fc9362f")}
          </div>
          <h1 className="mt-5 text-3xl leading-tight font-bold tracking-normal sm:text-4xl">
            {copy("copy.user.components.layout.route-error-page.d27b8e6d1494")}
          </h1>
          <p className="mt-4 text-sm leading-7 text-muted-foreground sm:text-base">
            {copy("copy.user.components.layout.route-error-page.cda8f4d19de2")}
          </p>
        </section>
      </main>
    </>
  )
}

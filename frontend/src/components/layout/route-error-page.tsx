import type { ErrorComponentProps } from "@tanstack/react-router"

import { PageTitle } from "@/components/common/page-title"

export function RouteErrorPage(_props: ErrorComponentProps) {
  return (
    <>
      <PageTitle>页面异常</PageTitle>
      <main className="flex min-h-svh items-center justify-center bg-background px-4 py-12 text-foreground">
        <section className="max-w-xl text-center">
          <div className="inline-flex items-center rounded-full border bg-card px-3 py-1 text-xs font-medium text-muted-foreground shadow-xs">
            页面异常
          </div>
          <h1 className="mt-5 text-3xl leading-tight font-bold tracking-normal sm:text-4xl">
            这页暂时没有正常加载
          </h1>
          <p className="mt-4 text-sm leading-7 text-muted-foreground sm:text-base">
            可能是网络请求、页面数据或前端状态出现了临时问题，请稍后再试。
          </p>
        </section>
      </main>
    </>
  )
}

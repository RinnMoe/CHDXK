import { Suspense } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider } from "react-router-dom"

import { AuthProvider } from "@/contexts/auth-context"
import { ThemeProvider } from "@/components/theme-provider"
import { queryClient } from "@/lib/query-client"
import { router } from "@/router"

function RouteFallback() {
  return (
    <div className="flex min-h-dvh items-center justify-center text-sm text-muted-foreground">
      加载中...
    </div>
  )
}

export function App() {
  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <Suspense fallback={<RouteFallback />}>
            <RouterProvider router={router} />
          </Suspense>
        </AuthProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

export default App

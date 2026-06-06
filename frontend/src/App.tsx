import { QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider } from "@tanstack/react-router"

import { AuthProvider } from "@/contexts/auth-context"
import { SystemSettingsLoader } from "@/components/system-settings-loader"
import { ThemeProvider } from "@/components/theme-provider"
import { queryClient } from "@/lib/query-client"
import { router } from "@/router"

export function App() {
  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <SystemSettingsLoader />
          <RouterProvider router={router} context={{ queryClient }} />
        </AuthProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

export default App

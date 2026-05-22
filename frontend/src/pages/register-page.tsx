import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { RegisterForm } from "@/components/auth/register-form"
import { PageShell } from "@/components/layout/page-shell"
import { useAuth } from "@/contexts/auth-context"

export function RegisterPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isLoading && user) {
      navigate("/", { replace: true })
    }
  }, [user, isLoading, navigate])

  if (isLoading) return null

  return (
    <PageShell>
      <div className="flex justify-center">
        <RegisterForm />
      </div>
    </PageShell>
  )
}

import { Link } from "react-router-dom"

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center px-4 max-w-5xl mx-auto">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <span className="text-lg">JCourse</span>
        </Link>
        <nav className="ml-8 flex gap-4 text-sm">
          <Link
            to="/courses"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            课程
          </Link>
        </nav>
      </div>
    </header>
  )
}

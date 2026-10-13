import { useState, type ReactNode } from "react"
import { Link, useNavigate } from "@tanstack/react-router"
import { RiSearchLine, RiRefreshLine } from "@remixicon/react"
import { copy } from "@/lib/copy"
import { PageShell } from "@/components/layout/page-shell"
import { PageTitle } from "@/components/common/page-title"
import { ReviewList } from "@/components/review/review-list"
import { HotCourseList } from "@/components/course/hot-course-list"
import { CourseCard } from "@/components/course/course-card"
import { CourseCompactCard } from "@/components/course/course-compact-card"
import { EnrollmentImportDialog } from "@/components/course/enrollment-import-dialog"
import {
  discoveryBatch,
  discoveryBatchSize,
  homeCategories,
} from "@/components/course/home-course-utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuth } from "@/contexts/auth-context"
import { useReviews } from "@/hooks/use-review"
import {
  useCourseDiscovery,
  useCourseFilters,
  useCourseReviewPrompts,
  useFollowedCourses,
  useHotCourses,
} from "@/hooks/use-course"

const text = (name: string) => copy(`copy.user.pages.home-page.${name}`)
const linkClass =
  "text-sm font-medium text-primary underline underline-offset-4 hover:text-primary/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-sm"

function HomeSection({
  title,
  action,
  children,
  className = "",
}: {
  title: string
  action?: ReactNode
  children: ReactNode
  className?: string
}) {
  return (
    <section className={`min-w-0 space-y-3 ${className}`} aria-label={title}>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="text-lg font-semibold">{title}</h2>
        {action}
      </div>
      {children}
    </section>
  )
}

function QueryFeedback({
  query,
}: {
  query: {
    isPending: boolean
    isError: boolean
    data: unknown
    refetch: () => unknown
  }
}) {
  if (query.isPending)
    return (
      <div className="space-y-3" role="status" aria-label={text("loading")}>
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-16 w-full" />
      </div>
    )
  if (!query.isError) return null
  return (
    <div
      role="status"
      className="flex flex-wrap items-center gap-2 rounded-md border p-3 text-sm text-muted-foreground"
    >
      <span>{text(query.data ? "refresh-error" : "load-error")}</span>
      <Button
        variant="outline"
        size="sm"
        onClick={() => {
          void query.refetch()
        }}
      >
        {text("retry")}
      </Button>
    </div>
  )
}

function HomeSearch() {
  const [q, setQ] = useState("")
  const navigate = useNavigate()
  const filters = useCourseFilters()
  return (
    <section className="space-y-4" aria-label={text("search-title")}>
      <div>
        <h1 className="text-2xl font-bold">{text("search-title")}</h1>
      </div>
      <form
        role="search"
        className="flex gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          void navigate({
            to: "/course",
            search: { q: q.trim() || undefined, page: 1 },
          })
        }}
      >
        <Input
          value={q}
          onChange={(event) => setQ(event.target.value)}
          placeholder={text("search-placeholder")}
          aria-label={text("search-placeholder")}
        />
        <Button type="submit">
          <RiSearchLine aria-hidden="true" />
          {text("search")}
        </Button>
      </form>
      <QueryFeedback query={filters} />
      <nav className="flex flex-wrap gap-2" aria-label={text("categories")}>
        {homeCategories(filters.data?.categories).map((category) => (
          <Button
            key={category.name}
            variant="outline"
            size="sm"
            className="h-auto min-h-8 max-w-full py-1 wrap-break-word whitespace-normal"
            asChild
          >
            <Link
              to="/course"
              search={{ categories: [category.name], page: 1 }}
            >
              {category.name}
            </Link>
          </Button>
        ))}
      </nav>
    </section>
  )
}

function DiscoverySection() {
  const query = useCourseDiscovery()
  const [batch, setBatch] = useState(0)
  const items = query.data?.items ?? []
  const visible = discoveryBatch(items, batch)
  return (
    <HomeSection
      title={text("discover")}
      className="lg:col-span-8"
      action={
        <div className="flex items-center gap-3">
          {items.length > discoveryBatchSize && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setBatch((value) => value + 1)}
            >
              <RiRefreshLine aria-hidden="true" />
              {text("next-batch")}
            </Button>
          )}
          <Link to="/course" className={linkClass}>
            {text("browse-all")}
          </Link>
        </div>
      }
    >
      <QueryFeedback query={query} />
      {query.data && (
        <>
          {items.length > 0 ? (
            <div
              className="grid gap-x-6 gap-y-2 sm:grid-cols-2"
              aria-live="polite"
            >
              {visible.map((course) => (
                <CourseCard
                  key={course.id}
                  course={course}
                  emptyRatingText={text("no-rating")}
                  alignRows
                />
              ))}
            </div>
          ) : (
            <p className="rounded-md border p-4 text-sm text-muted-foreground">
              {text("no-courses")}
            </p>
          )}
        </>
      )}
    </HomeSection>
  )
}

function ReviewPromptsSection() {
  const { user } = useAuth()
  const query = useCourseReviewPrompts(user?.id)
  return (
    <HomeSection
      title={text("share-experience")}
      className="lg:col-span-4 lg:border-l lg:pl-6"
      action={
        <Link
          to="/course/mine"
          search={{ type: "enrolled" }}
          className={linkClass}
        >
          {text("my-courses")}
        </Link>
      }
    >
      <QueryFeedback query={query} />
      {query.data &&
        (query.data.items.length > 0 ? (
          <div className="space-y-2">
            {query.data.items.map(({ course, semester }) => (
              <div key={course.id} className="rounded-md border p-3">
                <Link
                  to="/course/$courseID"
                  params={{ courseID: String(course.id) }}
                  className="font-medium hover:underline"
                >
                  {course.name}
                </Link>
                <p className="mt-1 text-sm text-muted-foreground">
                  {course.main_teacher.name} · {semester}
                </p>
                <Button variant="outline" size="sm" className="mt-3" asChild>
                  <Link
                    to="/course/$courseID/review/new"
                    params={{ courseID: String(course.id) }}
                  >
                    {text("write-review")}
                  </Link>
                </Button>
              </div>
            ))}
          </div>
        ) : query.data.has_enrollments ? (
          <p className="rounded-md border p-3 text-sm text-muted-foreground">
            {text("all-reviewed")}
          </p>
        ) : (
          <div className="space-y-3 rounded-md border p-3">
            <p className="text-sm text-muted-foreground">
              {text("no-enrollments")}
            </p>
            <div className="flex flex-wrap items-center gap-3">
              <EnrollmentImportDialog />
              <Link to="/course" className={linkClass}>
                {text("search-taken")}
              </Link>
            </div>
          </div>
        ))}
    </HomeSection>
  )
}

function LatestReviewsSection() {
  const query = useReviews({ page: 1, page_size: 6 })
  return (
    <HomeSection
      title={copy("copy.user.pages.home-page.45bdcac3bd65")}
      className="lg:col-span-8"
      action={
        <Link to="/review" className={linkClass}>
          {text("more")}
        </Link>
      }
    >
      <QueryFeedback query={query} />
      {query.data &&
        (query.data.items.length > 0 ? (
          <ReviewList reviews={query.data.items.slice(0, 6)} showCourse />
        ) : (
          <div className="flex flex-wrap items-center gap-3 text-sm">
            <p className="text-sm text-muted-foreground">
              {text("first-review")}
            </p>
            <Link
              to="/course/mine"
              search={{ type: "enrolled" }}
              className={linkClass}
            >
              {text("write-review")}
            </Link>
          </div>
        ))}
    </HomeSection>
  )
}

function HomeSidebar() {
  const hot = useHotCourses("week", 5)
  const followed = useFollowedCourses({ page: 1, page_size: 5 })
  return (
    <div className="min-w-0 space-y-6 lg:col-span-4 lg:border-l lg:pl-6">
      {(hot.isPending || hot.isError || !!hot.data?.items.length) && (
        <HomeSection
          title={copy("copy.user.pages.home-page.d6e3b43107c5")}
          action={
            <Link to="/course/hot" className={linkClass}>
              {text("more")}
            </Link>
          }
        >
          <QueryFeedback query={hot} />
          {!!hot.data?.items.length && (
            <HotCourseList
              period="week"
              limit={5}
              data={hot.data}
              isLoading={false}
              fetchData={false}
            />
          )}
        </HomeSection>
      )}
      <HomeSection
        title={copy("copy.user.pages.home-page.7be923dbf8a8")}
        action={
          <Link
            to="/course/mine"
            search={{ type: "followed" }}
            className={linkClass}
          >
            {text("more")}
          </Link>
        }
      >
        <QueryFeedback query={followed} />
        {followed.data &&
          (followed.data.items.length > 0 ? (
            followed.data.items.map((course) => (
              <CourseCompactCard key={course.id} course={course} />
            ))
          ) : (
            <p className="text-sm text-muted-foreground">
              {text("follow-description")}{" "}
              <Link to="/course" className={linkClass}>
                {text("discover")}
              </Link>
            </p>
          ))}
      </HomeSection>
    </div>
  )
}

export function HomePage() {
  return (
    <>
      <PageTitle />
      <PageShell>
        <div className="space-y-8">
          <HomeSearch />
          <div className="grid gap-8 lg:grid-cols-12">
            <DiscoverySection />
            <ReviewPromptsSection />
            <LatestReviewsSection />
            <HomeSidebar />
          </div>
        </div>
      </PageShell>
    </>
  )
}

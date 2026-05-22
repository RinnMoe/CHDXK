import { createBrowserRouter, Navigate } from "react-router-dom"
import { Layout } from "@/components/layout/layout"
import { CoursesPage } from "@/pages/courses-page"
import { CourseDetailPage } from "@/pages/course-detail-page"

export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { path: "/", element: <Navigate to="/courses" replace /> },
      { path: "/courses", element: <CoursesPage /> },
      { path: "/courses/:courseID", element: <CourseDetailPage /> },
    ],
  },
])

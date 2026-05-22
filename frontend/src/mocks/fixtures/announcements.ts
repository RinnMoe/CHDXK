import type { AnnouncementDTO } from "@/api/announcement"

const now = Date.now()

export const mockAnnouncements: AnnouncementDTO[] = [
  {
    id: 1,
    title: "新版评价系统上线",
    body: "为了提供更好的体验，我们重构了评价系统。如有问题请联系管理员。",
    priority: 10,
    created_at: new Date(now - 1000 * 60 * 60 * 24).toISOString(),
  },
  {
    id: 2,
    title: "积分系统说明",
    body: "发表点评、获得点赞、每日登录都可以获得积分。积分可以用于转账。",
    priority: 5,
    created_at: new Date(now - 1000 * 60 * 60 * 24 * 7).toISOString(),
  },
  {
    id: 3,
    title: "欢迎使用 JCourse",
    body: "JCourse 是一个课程评价平台，欢迎大家分享自己的选课心得。",
    priority: 1,
    created_at: new Date(now - 1000 * 60 * 60 * 24 * 30).toISOString(),
  },
]

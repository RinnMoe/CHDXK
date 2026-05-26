import type {
  CourseListItemDTO,
  CourseDetailDTO,
  CourseFilters,
} from "@/api/course"

const DEPARTMENTS = [
  "计算机科学与工程系",
  "数学科学学院",
  "物理与天文学院",
  "电子信息与电气工程学院",
  "跨学科复杂系统智能工程与可持续社会创新联合研究中心课程建设办公室",
  "外国语学院",
  "化学化工学院",
  "生命科学技术学院",
  "机械与动力工程学院",
]

const CATEGORIES = [
  "专业必修",
  "专业选修",
  "通识核心",
  "通识选修",
  "公共基础",
  "超长筛选项-面向真实产业场景的跨学院跨年级跨语种综合实践课程模块",
  "VeryLongCourseFilterCategoryWithoutSpacesForOverflowTestingAndLayoutValidation",
]

const TARGET_YEARS = [
  "大一",
  "大二",
  "大三",
  "大四",
  "研究生",
  "本硕博贯通培养项目高年级及跨专业联合选课学生",
]

const LANGUAGES = ["中文", "英文", "双语"]

const CREDITS = [1, 2, 3, 4]

export const MOCK_COURSE_SEMESTERS = [
  "2025-2026-1",
  "2024-2025-2",
  "2024-2025-1",
]

const TEACHER_NAMES = [
  "张伟",
  "王芳",
  "李娜",
  "刘洋",
  "陈静",
  "杨磊",
  "黄丽",
  "赵勇",
  "周敏",
  "吴杰",
  "徐峰",
  "孙颖",
  "马超",
  "朱琳",
  "胡军",
  "郭强",
  "何萍",
]

function randInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function pick<T>(arr: T[]): T {
  return arr[randInt(0, arr.length - 1)]
}

function pickMany<T>(arr: T[], n: number): T[] {
  const copy = [...arr]
  const result: T[] = []
  for (let i = 0; i < n && copy.length > 0; i++) {
    const idx = randInt(0, copy.length - 1)
    result.push(copy.splice(idx, 1)[0])
  }
  return result
}

let teacherIDSeed = 1
const teacherCache = new Map<
  string,
  { id: number; code: string; name: string; department: string; title?: string }
>()

function makeTeacher(name?: string) {
  const finalName = name ?? pick(TEACHER_NAMES)
  if (teacherCache.has(finalName)) return teacherCache.get(finalName)!
  const dept = pick(DEPARTMENTS)
  const teacher = {
    id: teacherIDSeed++,
    code: `T${String(teacherIDSeed).padStart(5, "0")}`,
    name: finalName,
    department: dept,
    title: pick(["教授", "副教授", "讲师", "助理教授"]),
  }
  teacherCache.set(finalName, teacher)
  return teacher
}

function makeDistribution(
  count: number
): [number, number, number, number, number] {
  if (count === 0) return [0, 0, 0, 0, 0]
  const dist: [number, number, number, number, number] = [0, 0, 0, 0, 0]
  for (let i = 0; i < count; i++) {
    dist[randInt(0, 4)]++
  }
  return dist
}

function avgFromDistribution(
  dist: [number, number, number, number, number]
): number {
  const total = dist.reduce((a, b) => a + b, 0)
  if (total === 0) return 0
  const sum = dist.reduce((s, c, i) => s + c * (i + 1), 0)
  return Math.round((sum / total) * 100) / 100
}

function ratingScore(avg: number, count: number): number {
  if (count === 0) return 0
  const globalAvg = 3.8
  const priorCount = 5
  return Math.round(((avg * count + priorCount * globalAvg) / (count + priorCount)) * 100) / 100
}

let courseIDSeed = 1

// Define code groups: most courses share a code with 1-3 other courses
const CODE_GROUPS: { code: string; name: string; credit: number }[] = [
  { code: "CS1001", name: "数据结构", credit: 4 },
  { code: "CS1002", name: "操作系统", credit: 4 },
  { code: "CS1003", name: "计算机网络", credit: 3 },
  { code: "CS1004", name: "编译原理", credit: 3 },
  { code: "MA1001", name: "高等数学", credit: 4 },
  { code: "MA1002", name: "线性代数", credit: 3 },
  { code: "MA2001", name: "概率论与数理统计", credit: 3 },
  { code: "MA2002", name: "离散数学", credit: 3 },
  { code: "PH1001", name: "大学物理", credit: 4 },
  { code: "PH2001", name: "理论力学", credit: 3 },
  { code: "PH2002", name: "电磁学", credit: 3 },
  { code: "PH2003", name: "量子力学", credit: 3 },
  { code: "EN1001", name: "大学英语", credit: 3 },
  { code: "EN2001", name: "学术英语写作", credit: 2 },
  { code: "CH1001", name: "有机化学", credit: 3 },
  { code: "CH1002", name: "无机化学", credit: 3 },
  { code: "BI1001", name: "细胞生物学", credit: 3 },
  { code: "BI2001", name: "分子生物学", credit: 3 },
  { code: "ME1001", name: "机械设计基础", credit: 3 },
  { code: "ME2001", name: "热力学", credit: 3 },
  { code: "AI1001", name: "人工智能导论", credit: 3 },
  { code: "AI2001", name: "机器学习", credit: 3 },
  { code: "AI2002", name: "深度学习", credit: 3 },
  { code: "CS2001", name: "数据库系统", credit: 3 },
  { code: "CS2002", name: "软件工程", credit: 3 },
  { code: "CS2003", name: "算法分析与设计", credit: 3 },
  { code: "CS2004", name: "计算机组成原理", credit: 3 },
  { code: "EE1001", name: "信号与系统", credit: 3 },
  { code: "EE1002", name: "数字电路", credit: 3 },
  { code: "EE1003", name: "模拟电路", credit: 3 },
]

function makeCourse(): CourseListItemDTO {
  const group = pick(CODE_GROUPS)
  const count = randInt(0, 50)
  const distribution = makeDistribution(count)
  const avg = avgFromDistribution(distribution)
  const mainTeacher = makeTeacher()
  return {
    id: courseIDSeed++,
    code: group.code,
    name: group.name,
    credit: group.credit,
    department: mainTeacher.department,
    language: pick(LANGUAGES),
    target_years: pickMany(TARGET_YEARS, randInt(1, 2)),
    categories: pickMany(CATEGORIES, randInt(1, 2)),
    main_teacher: mainTeacher,
    rating: {
      count,
      avg,
      score: ratingScore(avg, count),
      distribution,
    },
  }
}

export const mockCourses: CourseListItemDTO[] = Array.from({ length: 64 }, () =>
  makeCourse()
)

mockCourses[0] = {
  ...mockCourses[0],
  name: "面向复杂真实世界系统的超大规模分布式数据库架构设计与性能调优实践",
  department: "电子信息与电气工程学院",
  language: "中文及英文双语研讨",
  categories: ["跨学科综合实践课程", "研究型专业选修模块"],
  main_teacher: makeTeacher("欧阳明远清和"),
}

mockCourses[1] = {
  ...mockCourses[1],
  name: "人工智能安全、可信机器学习与大模型治理专题前沿导论",
  department: "计算机科学与工程系",
  categories: ["通识核心-科技伦理与社会", "专业方向拓展"],
  main_teacher: makeTeacher("司徒嘉言"),
}

mockCourses[2] = {
  ...mockCourses[2],
  name: "计算社会科学中的因果推断、网络实验与高维数据分析方法",
  department: "数学科学学院",
  language: "全英文授课与中文讨论",
  categories: ["方法论强化训练", "通识选修-社会科学"],
  main_teacher: makeTeacher("Alexander Christopher Johnson-Smith"),
}

// Courses with no reviews for testing zero-state UI
const noRatingDistribution: [number, number, number, number, number] = [
  0, 0, 0, 0, 0,
]
mockCourses.push(
  {
    id: courseIDSeed++,
    code: "CS0101",
    name: "量子计算导论",
    credit: 3,
    department: "计算机科学与工程系",
    language: "中文",
    target_years: ["大三", "大四"],
    categories: ["专业选修"],
    main_teacher: makeTeacher("钱学"),
    rating: { count: 0, avg: 0, score: 0, distribution: noRatingDistribution },
  },
  {
    id: courseIDSeed++,
    code: "MA0087",
    name: "拓扑学基础",
    credit: 2,
    department: "数学科学学院",
    language: "中文",
    target_years: ["研究生"],
    categories: ["专业必修"],
    main_teacher: makeTeacher("孙理"),
    rating: { count: 0, avg: 0, score: 0, distribution: noRatingDistribution },
  },
  {
    id: courseIDSeed++,
    code: "PH0042",
    name: "天体物理",
    credit: 3,
    department: "物理与天文学院",
    language: "英文",
    target_years: ["大三", "研究生"],
    categories: ["专业选修", "通识选修"],
    main_teacher: makeTeacher("李星"),
    rating: { count: 0, avg: 0, score: 0, distribution: noRatingDistribution },
  }
)

export function makeCourseDetail(course: CourseListItemDTO): CourseDetailDTO {
  const byRatingScore = (a: CourseListItemDTO, b: CourseListItemDTO) =>
    b.rating.score - a.rating.score ||
    b.rating.count - a.rating.count ||
    b.rating.avg - a.rating.avg ||
    a.code.localeCompare(b.code)
  const sameCode = mockCourses
    .filter((c) => c.code === course.code && c.id !== course.id)
    .sort(byRatingScore)
  const sameTeacher = mockCourses
    .filter(
      (c) => c.main_teacher.id === course.main_teacher.id && c.id !== course.id
    )
    .sort(byRatingScore)
  const currentTeacherGroup = [
    course.main_teacher,
    makeTeacher(pick(TEACHER_NAMES)),
  ]

  return {
    id: course.id,
    code: course.code,
    name: course.name,
    credit: course.credit,
    department: course.department,
    last_semester: "2025-2026-1",
    language: course.language,
    target_years: course.target_years,
    categories: course.categories,
    main_teacher: course.main_teacher,
    teacher_group: currentTeacherGroup,
    rating: course.rating,
    notification_level: 0,
    offered_courses: [
      {
        semester: "2025-2026-1",
        language: course.language,
        target_years: course.target_years,
        categories: course.categories,
      },
      {
        semester: "2024-2025-2",
        language: course.language,
        target_years: course.target_years,
        categories: course.categories,
      },
    ],
    same_code_courses: sameCode,
    same_teacher_courses: sameTeacher,
  }
}

export function makeCourseFilters(): CourseFilters {
  const counts = (arr: string[]) =>
    arr.map((name) => ({
      name,
      count: randInt(5, 60),
    }))
  return {
    credits: CREDITS.map((c) => ({ name: String(c), count: randInt(5, 30) })),
    departments: counts(DEPARTMENTS),
    categories: counts(CATEGORIES),
    target_years: counts(TARGET_YEARS),
    languages: counts(["中文", "英文", "双语"]),
    semesters: counts(MOCK_COURSE_SEMESTERS),
  }
}

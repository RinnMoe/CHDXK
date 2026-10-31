# Backend Redis Keys

This document lists app-owned Redis keys used by the backend. Keys should follow:

```text
jcourse:{domain}:{part...}
```

Repository code should build keys with the local `redisKey` helper in `internal/infrastructure/repository/cache.go`; cache keys continue to go through `cacheKey`, which delegates to the same helper. Repository JSON caches are best-effort and use the default TTL of 30 minutes.

Redis keys owned internally by third-party libraries can have library-defined formats. The session store is configured with the `jcourse:session:` prefix; Asynq keys are managed by Asynq.

## Repository JSON Caches

| Key pattern | Owner | Cached value | Invalidated by |
| --- | --- | --- | --- |
| `jcourse:course:{course_id}` | `CourseRepository.Get` | `course.Course` | Review create/update/delete for the course |
| `jcourse:course:{course_id}:detail` | `CourseRepository.GetDetail` | `course.CourseDetailView` | Review create/update/delete for the course |
| `jcourse:course:{course_id}:offered:{semester}` | `CourseRepository.OfferedCourseExists` | `bool` | TTL only |
| `jcourse:course:filters` | `CourseRepository.GetFilters` | `course.CourseFilters` | Review create/delete |
| `jcourse:course:hot:{period}:{period_key}:{limit}` | `GormCourseHotRepository.Top` | `[]course.HotCourseRank` | `GormCourseHotRepository.AddScore`, by pattern `jcourse:course:hot:*:*` |
| `jcourse:course_notification:{user_id}:{course_id}:level` | `CourseNotificationRepository.GetLevel` | `course.NotificationLevel` | `CourseNotificationRepository.SetLevel` |
| `jcourse:course_notification:{user_id}:level:{level}:courses` | `CourseNotificationRepository.GetCoursesByLevel` | `[]int` course IDs | `CourseNotificationRepository.SetLevel`, by pattern `jcourse:course_notification:{user_id}:level:*:courses` |
| `jcourse:point:{user_id}:sum` | `PointRepository.SumByUser` | `int` point balance | `PointRepository.CreateTransfer` for sender and recipient |
| `jcourse:review:{review_id}` | `ReviewRepository.Get` | `review.Review` | Review create/update/delete, moderator remark update |
| `jcourse:review:{review_id}:view` | `ReviewRepository.GetByID` | `review.ReviewView` | Review create/update/delete, moderator remark update, vote save/delete |
| `jcourse:review:course:{course_id}:filters` | `ReviewRepository.GetCourseFilters` | `review.ReviewFilters` | Review create/update/delete for the course |
| `jcourse:review:course:{course_id}:trend` | `ReviewRepository.GetCourseTrend` | `[]review.ReviewTrendItem` | Review create/update/delete for the course |
| `jcourse:review_vote:{review_id}:{user_id}` | `ReviewVoteRepository.FindByReviewAndUser` | `review.Vote` | `ReviewVoteRepository.Save`, `ReviewVoteRepository.Delete` |
| `jcourse:canteen:home:v2` | `CanteenStoreQueryRepository.GetHome` | `canteen.HomeView` with hero/feeds/hot dishes | Any canteen review/store/hall/dish write, by pattern `jcourse:canteen:*` |
| `jcourse:canteen:store:{store_id}:detail:v2:{viewer_id}` | `CanteenStoreQueryRepository.GetStoreDetail` | `canteen.StoreDetail` (viewer-specific my summary, dish covers) | Any canteen review/store/hall/dish write, by pattern `jcourse:canteen:*` |
| `jcourse:canteen:rankings:v2:{campus}:{hall_id}:{sort}:{page}:{page_size}` | `CanteenStoreQueryRepository.ListRankings` | `canteenRankingCache` with confidence/tags/dish counts | Any canteen review/store/hall/dish write, by pattern `jcourse:canteen:*` |
| `jcourse:site_daily_stat:{stat_date}` | `SiteDailyStatRepository.GetByDate` | `stat.DailyStatView` | `SiteDailyStatRepository.Upsert` |
| `jcourse:teacher:{teacher_id}` | `TeacherRepository.GetByID` | `teacher.TeacherView` | TTL only |
| `jcourse:teacher:filters` | `TeacherRepository.GetFilters` | `teacher.TeacherFilters` | TTL only |
| `jcourse:user:{user_id}` | `UserRepository.FindByID` | `auth.User` authorization projection | `UserRepository.Update` |

`{period}` is `week` or `month`. `{period_key}` uses `HotCoursePeriodKey`: weekly keys are `YYYY-WW`, monthly keys are `YYYY-MM`. `{stat_date}` uses Go `time.DateOnly` format, `YYYY-MM-DD`.

## Other Redis Keys

| Key pattern | Owner | Data structure | TTL |
| --- | --- | --- | --- |
| `jcourse:course:hot:week:{period_key}` | `CourseHotRepository` | Redis sorted set, member is course ID, score is hot score | No TTL currently set |
| `jcourse:course:hot:month:{period_key}` | `CourseHotRepository` | Redis sorted set, member is course ID, score is hot score | No TTL currently set |
| `jcourse:session:{session_id}` | Gin session Redis store | Serialized session payload | Session max age from session config |
| `jcourse:oidc-refresh:{sha256(refresh_token)}` | OIDC refresh coordinator | Rotated access/refresh credentials, server-only | 2 minutes |
| `jcourse:oidc-refresh:{sha256(refresh_token)}:lock` | OIDC refresh coordinator | Random lock owner, compare-and-delete release | 30 seconds |
| `jcourse:oidc-refresh:{sha256(refresh_token)}:consumed` | OIDC refresh coordinator | Single-use grant consumption marker | Configured session max age |
| `jcourse:session-revocation:{sha256(session_id)}` | Session save gate | Logout/session-ID rotation marker; rejects stale writes | Configured session max age |
| `jcourse:session-revocation:{sha256(session_id)}:lock` | Session save gate | Random owner; serializes save and deletion across API instances | 30 seconds |

OIDC access/ID/refresh tokens, access expiry and absolute local session deadline are stored only in the server-side session payload; state, PKCE verifier and nonce are stored during authorization. The short-lived rotation result lets concurrent API instances share one refresh without replaying a grant. The consumption marker prevents replay after a lost response, process crash or expired result cache. These keys contain credentials: apply the same access controls as session data; never log their payloads. No local registration, password-reset, verification-code, or login-lockout Redis keys are owned by the application.

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
| `jcourse:site_daily_stat:{stat_date}` | `SiteDailyStatRepository.GetByDate` | `stat.DailyStatView` | `SiteDailyStatRepository.Upsert` |
| `jcourse:teacher:{teacher_id}` | `TeacherRepository.GetByID` | `teacher.TeacherView` | TTL only |
| `jcourse:teacher:filters` | `TeacherRepository.GetFilters` | `teacher.TeacherFilters` | TTL only |
| `jcourse:user:{user_id}` | `UserRepository.FindByID` | `auth.User` | `UserRepository.Update` |
| `jcourse:account:{user_id}` | `AccountRepository.FindByID` | `account.Account` | Account create/update/touch last seen, user update |
| `jcourse:account:email:{email}` | `AccountRepository.FindByEmail` | `account.Account` | Account create/update/touch last seen, user update |

`{period}` is `week` or `month`. `{period_key}` uses `HotCoursePeriodKey`: weekly keys are `YYYY-WW`, monthly keys are `YYYY-MM`. `{stat_date}` uses Go `time.DateOnly` format, `YYYY-MM-DD`.

## Other Redis Keys

| Key pattern | Owner | Data structure | TTL |
| --- | --- | --- | --- |
| `jcourse:auth:{prefix}:code:{lower_email}` | `VerificationCodeRepository.Save/Get/Delete` | String `{code}|{expires_at_unix}` | Verification code TTL from auth config |
| `jcourse:auth:{prefix}:code_cooldown:{lower_email}` | `VerificationCodeRepository.ReserveSend` | String marker | Send interval from auth config |
| `jcourse:auth:login_attempts:{lower_email}` | `LoginAttemptRepository` | Integer counter | Login lockout duration from auth config, set when counter is first created |
| `jcourse:course:hot:week:{period_key}` | `CourseHotRepository` | Redis sorted set, member is course ID, score is hot score | No TTL currently set |
| `jcourse:course:hot:month:{period_key}` | `CourseHotRepository` | Redis sorted set, member is course ID, score is hot score | No TTL currently set |
| `jcourse:session:{session_id}` | Gin session Redis store | Serialized session payload | Session max age from session config |

`VerificationCodeRepository` currently uses `register` and `reset` as prefixes.

# Backend Cache Keys

This document lists Redis keys used by the backend. Repository JSON caches use the shared `cacheKey` helper and follow:

```text
jcourse:{domain}:{entity_id_or_scope...}
```

Repository JSON caches are best-effort and use the default TTL of 30 minutes.

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

These keys predate the repository JSON cache helper and keep their existing prefixes and data structures.

| Key pattern | Owner | Data structure | TTL |
| --- | --- | --- | --- |
| `auth:{prefix}_code:{lower_email}` | `VerificationCodeRepository.Save/Get/Delete` | String `{code}|{expires_at_unix}` | Verification code TTL from auth config |
| `auth:{prefix}_code_cooldown:{lower_email}` | `VerificationCodeRepository.ReserveSend` | String marker | Send interval from auth config |
| `auth:login_attempts:{lower_email}` | `LoginAttemptRepository` | Integer counter | Login lockout duration from auth config, set when counter is first created |
| `course:hot:week:{period_key}` | `CourseHotRepository` | Redis sorted set, member is course ID, score is hot score | No TTL currently set |
| `course:hot:month:{period_key}` | `CourseHotRepository` | Redis sorted set, member is course ID, score is hot score | No TTL currently set |

`VerificationCodeRepository` currently uses `register` and `reset` as prefixes.

package apperr

import "net/http"

type AppError struct {
	Msg        string
	StatusCode int
}

func New(msg string, statusCode int) *AppError {
	return &AppError{Msg: msg, StatusCode: statusCode}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func BadRequest(msg string) *AppError {
	return New(msg, http.StatusBadRequest)
}

func Unauthorized(msg string) *AppError {
	return New(msg, http.StatusUnauthorized)
}

func Forbidden(msg string) *AppError {
	return New(msg, http.StatusForbidden)
}

func NotFound(msg string) *AppError {
	return New(msg, http.StatusNotFound)
}

func Conflict(msg string) *AppError {
	return New(msg, http.StatusConflict)
}

func TooManyRequests(msg string) *AppError {
	return New(msg, http.StatusTooManyRequests)
}

func ServiceUnavailable(msg string) *AppError {
	return New(msg, http.StatusServiceUnavailable)
}

var (
	ErrUnauthorized = Unauthorized("未登录")

	ErrUserAlreadyExists  = Conflict("用户已存在")
	ErrUserNotFound       = NotFound("用户不存在")
	ErrEmailNotAllowed    = Forbidden("邮箱不允许注册")
	ErrPasswordRequired   = BadRequest("密码不能为空")
	ErrInvalidCredentials = Unauthorized("邮箱或密码错误")
	ErrLoginLocked        = TooManyRequests("登录失败次数过多，请稍后再试")

	ErrVerificationSendTooSoon = TooManyRequests("验证码发送过于频繁")
	ErrVerificationCodeInvalid = BadRequest("验证码无效")

	ErrUserSuspended          = Forbidden("用户已被封禁")
	ErrApiKeyNameRequired     = BadRequest("API Key 名称不能为空")
	ErrApiKeyLimitExceeded    = BadRequest("API Key 数量已达上限")
	ErrApiKeyNotFound         = NotFound("API Key 不存在")
	ErrInvalidApiKey          = Unauthorized("API Key 无效")
	ErrInvalidApiKeyRole      = BadRequest("API Key 角色无效")
	ErrCannotSuspendAdmin     = Forbidden("不能封禁管理员用户")
	ErrCannotOperateSelf      = Forbidden("不能操作自己")
	ErrRequireSuperAdmin      = Forbidden("需要超级管理员权限")
	ErrCannotModifySuperAdmin = Forbidden("不能修改超级管理员权限")

	ErrCourseNotFound           = NotFound("课程不存在")
	ErrTeacherNotFound          = NotFound("教师不存在")
	ErrSemesterRequired         = BadRequest("学期不能为空")
	ErrOfferedCourseNotFound    = BadRequest("该学期未开设此课程")
	ErrInvalidSyncSemester      = BadRequest("学期无效")
	ErrEnrollmentSyncDisabled   = ServiceUnavailable("选课同步功能已关闭")
	ErrInvalidHotCoursePeriod   = BadRequest("热门课程周期无效")
	ErrInvalidHotCourseActivity = BadRequest("热门课程活动类型无效")

	ErrReviewNotFound           = NotFound("点评不存在")
	ErrOfferedCourseMissing     = BadRequest("该学期未开设此课程")
	ErrUserCannotCreateReview   = Forbidden("无权创建点评")
	ErrUserCannotUpdateReview   = Forbidden("无权更新点评")
	ErrUserCannotModerateReview = Forbidden("无权更新管理员备注")
	ErrUserCannotDeleteReview   = Forbidden("无权删除点评")
	ErrContentSensitive         = BadRequest("点评内容包含敏感信息")
	ErrSimilarContentDetected   = BadRequest("短时间内相似点评过多")
	ErrSameCourseSpam           = BadRequest("短时间内同一课程点评过多")
	ErrDailyVoteLimitReached    = TooManyRequests("今日投票次数已达上限")
	ErrInvalidVoteType          = BadRequest("投票类型无效")
	ErrInvalidRating            = BadRequest("评分无效")
	ErrReviewContentEmpty       = BadRequest("点评内容不能为空")
	ErrReviewContentTooLong     = BadRequest("点评内容不能超过 9681 个字符")

	ErrInsufficientPointBalance          = Conflict("积分余额不足")
	ErrPointTransferInvalidAmount        = BadRequest("转账积分必须大于 0")
	ErrPointTransferInvalidFeePayer      = BadRequest("手续费承担方无效")
	ErrPointTransferSelf                 = BadRequest("不能给自己转账")
	ErrPointTransferRecipientAmountSmall = BadRequest("扣除手续费后到账积分过少")
	ErrPointTransferRecipientNotFound    = NotFound("收款用户不存在")
	ErrPointUserNotFound                 = NotFound("用户不存在")

	ErrInvalidCurrentSemester = BadRequest("当前学期无效")
	ErrInvalidDateRange       = BadRequest("日期范围无效")
	ErrInvalidAuditTimeRange  = BadRequest("审计日志时间范围无效")

	ErrJAccountDisabled         = ServiceUnavailable("JAccount 选课同步功能已关闭")
	ErrJAccountSemesterRequired = BadRequest("学期不能为空")
)

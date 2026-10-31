package canteen

import (
	"net/http"

	"jcourse/pkg/apperr"
)

var (
	ErrHallNotFound = apperr.NotFound("食堂不存在")
	ErrHallInvalid  = apperr.BadRequest("食堂信息不完整")

	ErrStoreNotFound  = apperr.NotFound("店铺不存在")
	ErrStoreNotActive = apperr.Conflict("店铺当前不可评价")
	ErrStoreMerged    = apperr.Conflict("店铺已合并到其他店铺")

	ErrDishNotFound      = apperr.NotFound("菜品不存在")
	ErrDishStoreMismatch = apperr.BadRequest("菜品不属于该店铺")
	ErrDishMerged        = apperr.Conflict("菜品已合并到其他菜品")

	ErrReviewNotFound   = apperr.NotFound("评价不存在")
	ErrReviewConflict   = apperr.Conflict("评价已被修改，请刷新后重试")
	ErrReviewNotVisible = apperr.Forbidden("评价不可见")
	ErrCannotEditReview = apperr.Forbidden("只能修改自己的评价")
	ErrCannotModerate   = apperr.Forbidden("没有权限操作该评价")

	ErrRatingOutOfRange = apperr.BadRequest("评分必须在 1 到 5 之间")
	ErrContentTooLong   = apperr.BadRequest("评价内容不能超过 2000 个字符")
	ErrTooManyTags      = apperr.BadRequest("体验标签最多选择 5 个")
	ErrInvalidTag       = apperr.BadRequest("体验标签无效")
	ErrDuplicateTag     = apperr.BadRequest("体验标签不能重复")

	ErrDailyReviewLimit = apperr.TooManyRequests("今日评价次数已达上限")
	ErrTooFrequent      = apperr.TooManyRequests("评价过于频繁，请稍后再试")
	ErrSimilarReview    = apperr.BadRequest("短时间内相似评价过多")
	ErrVoteOwnReview    = apperr.Forbidden("不能给自己的评价投票")
	ErrInvalidVoteType  = apperr.BadRequest("投票类型无效")

	ErrMediaNotFound     = apperr.NotFound("图片不存在")
	ErrMediaNotOwned     = apperr.Forbidden("只能操作自己上传的图片")
	ErrMediaState        = apperr.Conflict("图片当前状态不允许该操作")
	ErrMediaLimit        = apperr.BadRequest("每条评价最多关联 3 张图片")
	ErrMediaTooManyDaily = apperr.TooManyRequests("今日上传次数已达上限")
	ErrMediaTooLarge     = apperr.New("图片大小超出限制", http.StatusRequestEntityTooLarge)
	ErrMediaType         = apperr.New("仅支持 JPEG/PNG 图片", http.StatusUnsupportedMediaType)
	ErrMediaDimension    = apperr.BadRequest("图片尺寸超出限制")

	ErrReportNotFound  = apperr.NotFound("举报不存在")
	ErrReportDuplicate = apperr.Conflict("已举报过该内容")
	ErrReportReason    = apperr.BadRequest("举报原因无效")
)

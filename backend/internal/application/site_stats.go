package application

import (
	"context"
	"errors"
	"time"

	"jcourse/internal/domain/stat"
)

const dateLayout = "2006-01-02"

var ErrInvalidDateRange = errors.New("invalid date range")

type SiteDailyStatListFilter struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

type SiteStatsCommandService struct {
	collector stat.DailyStatCollector
	repo      stat.DailyStatCommandRepository
	loc       *time.Location
}

func NewSiteStatsCommandService(
	collector stat.DailyStatCollector,
	repo stat.DailyStatCommandRepository,
) *SiteStatsCommandService {
	return &SiteStatsCommandService{collector: collector, repo: repo, loc: mustStatsLocation()}
}

func (s *SiteStatsCommandService) CollectYesterday(ctx context.Context) (*SiteDailyStatDTO, error) {
	return s.CollectDaily(ctx, time.Now().In(s.loc).AddDate(0, 0, -1))
}

func (s *SiteStatsCommandService) CollectDailyByDateString(ctx context.Context, statDate string) (*SiteDailyStatDTO, error) {
	if statDate == "" {
		return s.CollectYesterday(ctx)
	}
	date, err := s.parseDate(statDate)
	if err != nil {
		return nil, err
	}
	return s.CollectDaily(ctx, date)
}

func (s *SiteStatsCommandService) CollectDaily(ctx context.Context, statDate time.Time) (*SiteDailyStatDTO, error) {
	date := s.dateOnly(statDate)
	periodStart := date
	periodEnd := periodStart.AddDate(0, 0, 1)

	metrics, err := s.collector.Collect(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(s.loc)
	daily := &stat.DailyStat{
		StatDate:    date,
		Metrics:     metrics,
		GeneratedAt: now,
		UpdatedAt:   now,
	}
	if err := s.repo.Upsert(ctx, daily); err != nil {
		return nil, err
	}
	dto := newSiteDailyStatDTO(daily)
	return &dto, nil
}

type SiteStatsQueryService struct {
	query stat.DailyStatQuery
	loc   *time.Location
}

func NewSiteStatsQueryService(query stat.DailyStatQuery) *SiteStatsQueryService {
	return &SiteStatsQueryService{query: query, loc: mustStatsLocation()}
}

func (s *SiteStatsQueryService) GetYesterday(ctx context.Context) (*SiteDailyStatDTO, error) {
	yesterday := s.dateOnly(time.Now().In(s.loc).AddDate(0, 0, -1))
	stat, err := s.query.GetByDate(ctx, yesterday)
	if err != nil {
		return nil, err
	}
	dto := newSiteDailyStatViewDTO(stat)
	return &dto, nil
}

func (s *SiteStatsQueryService) ListDaily(ctx context.Context, f SiteDailyStatListFilter) (*PaginatedResult[SiteDailyStatDTO], error) {
	startDate, endDate, err := s.parseDateRange(f.StartDate, f.EndDate)
	if err != nil {
		return nil, err
	}

	items, total, err := s.query.FindByDateRange(ctx, stat.DailyStatFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      f.Page,
		PageSize:  f.PageSize,
	})
	if err != nil {
		return nil, err
	}

	dtos := make([]SiteDailyStatDTO, len(items))
	for i, item := range items {
		dtos[i] = newSiteDailyStatViewDTO(&item)
	}

	return &PaginatedResult[SiteDailyStatDTO]{
		Items:    dtos,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func (s *SiteStatsQueryService) parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	start, err := s.parseDate(startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := s.parseDate(endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	return start, end, nil
}

func (s *SiteStatsCommandService) parseDate(value string) (time.Time, error) {
	date, err := time.ParseInLocation(dateLayout, value, s.loc)
	if err != nil {
		return time.Time{}, err
	}
	return s.dateOnly(date), nil
}

func (s *SiteStatsCommandService) dateOnly(value time.Time) time.Time {
	v := value.In(s.loc)
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, s.loc)
}

func (s *SiteStatsQueryService) parseDate(value string) (time.Time, error) {
	date, err := time.ParseInLocation(dateLayout, value, s.loc)
	if err != nil {
		return time.Time{}, err
	}
	return s.dateOnly(date), nil
}

func (s *SiteStatsQueryService) dateOnly(value time.Time) time.Time {
	v := value.In(s.loc)
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, s.loc)
}

func mustStatsLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	return loc
}

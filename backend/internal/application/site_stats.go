package application

import (
	"context"
	"strings"
	"time"

	"jcourse/internal/domain/stat"
)

var ErrInvalidDateRange = stat.ErrInvalidDateRange

type SiteDailyStatListFilter struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

type SiteStatsCommandService struct {
	daily *stat.DailyStatService
}

func NewSiteStatsCommandService(
	collector stat.DailyStatCollector,
	repo stat.DailyStatCommandRepository,
	config stat.Config,
) *SiteStatsCommandService {
	return &SiteStatsCommandService{daily: stat.NewDailyStatService(collector, repo, mustStatsLocation(config))}
}

func (s *SiteStatsCommandService) CollectDailyByDateString(ctx context.Context, statDate string) (*SiteDailyStatDTO, error) {
	daily, err := s.daily.CollectDailyByDateString(ctx, statDate)
	if err != nil {
		return nil, err
	}
	dto := newSiteDailyStatDTO(daily)
	return &dto, nil
}

func (s *SiteStatsCommandService) CollectDaily(ctx context.Context, statDate time.Time) (*SiteDailyStatDTO, error) {
	daily, err := s.daily.CollectDaily(ctx, statDate)
	if err != nil {
		return nil, err
	}
	dto := newSiteDailyStatDTO(daily)
	return &dto, nil
}

type SiteStatsQueryService struct {
	query    stat.DailyStatQuery
	calendar stat.Calendar
}

func NewSiteStatsQueryService(query stat.DailyStatQuery, config stat.Config) *SiteStatsQueryService {
	return &SiteStatsQueryService{query: query, calendar: stat.NewCalendar(mustStatsLocation(config))}
}

func (s *SiteStatsQueryService) GetByDateString(ctx context.Context, dateStr string) (*SiteDailyStatDTO, error) {
	date, err := s.calendar.ParseDate(dateStr)
	if err != nil {
		return nil, err
	}
	stat, err := s.query.GetByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if stat == nil {
		return nil, nil
	}
	dto := newSiteDailyStatViewDTO(stat)
	return &dto, nil
}

func (s *SiteStatsQueryService) ListDaily(ctx context.Context, f SiteDailyStatListFilter) (*PaginatedResult[SiteDailyStatDTO], error) {
	startDate, endDate, err := s.calendar.ParseDateRange(f.StartDate, f.EndDate)
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

func mustStatsLocation(config stat.Config) *time.Location {
	if strings.TrimSpace(config.Timezone) == "" {
		config.Timezone = stat.DefaultConfig.Timezone
	}
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		panic(err)
	}
	return loc
}

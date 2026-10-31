package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/canteen"
)

type canteenHallRow struct {
	ID           int    `gorm:"column:id"`
	CampusKey    string `gorm:"column:campus_key"`
	Name         string `gorm:"column:name"`
	ShortName    string `gorm:"column:short_name"`
	Address      string `gorm:"column:address"`
	Status       string `gorm:"column:status"`
	DisplayOrder int    `gorm:"column:display_order"`
	StoreCount   int    `gorm:"column:store_count"`
}

type canteenStoreListRow struct {
	ID               int            `gorm:"column:id"`
	HallID           int            `gorm:"column:hall_id"`
	HallName         string         `gorm:"column:hall_name"`
	CampusKey        string         `gorm:"column:campus_key"`
	Name             string         `gorm:"column:name"`
	Floor            string         `gorm:"column:floor"`
	Categories       pq.StringArray `gorm:"column:categories;type:text[]"`
	Status           string         `gorm:"column:status"`
	CoverMediaID     *int           `gorm:"column:cover_media_id"`
	RatingCount      int            `gorm:"column:rating_count"`
	RatingAvg        float64        `gorm:"column:rating_avg"`
	RatingScore      float64        `gorm:"column:rating_score"`
	TasteAvg         float64        `gorm:"column:taste_avg"`
	ValueAvg         float64        `gorm:"column:value_avg"`
	QueueAvg         float64        `gorm:"column:queue_avg"`
	HygieneAvg       float64        `gorm:"column:hygiene_avg"`
	ServiceAvg       float64        `gorm:"column:service_avg"`
	ReviewEventCount int            `gorm:"column:review_event_count"`
	LastReviewAt     *time.Time     `gorm:"column:last_review_at"`
}

type canteenReviewRow struct {
	ID           int            `gorm:"column:id"`
	StoreID      int            `gorm:"column:store_id"`
	StoreName    string         `gorm:"column:store_name"`
	UserID       int            `gorm:"column:user_id"`
	VisitedAt    time.Time      `gorm:"column:visited_at"`
	Taste        int            `gorm:"column:taste"`
	Value        int            `gorm:"column:value"`
	Queue        int            `gorm:"column:queue"`
	Hygiene      int            `gorm:"column:hygiene"`
	Service      int            `gorm:"column:service"`
	OverallScore float64        `gorm:"column:overall_score"`
	Content      string         `gorm:"column:content"`
	Tags         pq.StringArray `gorm:"column:tags;type:text[]"`
	Status       string         `gorm:"column:status"`
	LikeCount    int            `gorm:"column:like_count"`
	DislikeCount int            `gorm:"column:dislike_count"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
}

func newCanteenStoreListItem(row canteenStoreListRow) canteen.StoreListItem {
	return canteen.StoreListItem{
		ID:               row.ID,
		HallID:           row.HallID,
		HallName:         row.HallName,
		CampusKey:        row.CampusKey,
		Name:             row.Name,
		Floor:            row.Floor,
		Categories:       []string(row.Categories),
		Status:           canteen.StoreStatus(row.Status),
		CoverMediaID:     row.CoverMediaID,
		RatingCount:      row.RatingCount,
		RatingAvg:        row.RatingAvg,
		RatingScore:      row.RatingScore,
		ReviewEventCount: row.ReviewEventCount,
		LastReviewAt:     row.LastReviewAt,
	}
}

func newCanteenReviewView(row canteenReviewRow) canteen.ReviewView {
	return canteen.ReviewView{
		ID:        row.ID,
		StoreID:   row.StoreID,
		StoreName: row.StoreName,
		UserID:    row.UserID,
		VisitedAt: row.VisitedAt,
		Dimensions: canteen.Dimensions{
			Taste:   row.Taste,
			Value:   row.Value,
			Queue:   row.Queue,
			Hygiene: row.Hygiene,
			Service: row.Service,
		},
		OverallScore: row.OverallScore,
		Content:      row.Content,
		Tags:         []string(row.Tags),
		Status:       canteen.ReviewStatus(row.Status),
		LikeCount:    row.LikeCount,
		DislikeCount: row.DislikeCount,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

const canteenStoreListSelect = `
	s.id AS id,
	s.hall_id AS hall_id,
	h.name AS hall_name,
	h.campus_key AS campus_key,
	s.name AS name,
	s.floor AS floor,
	s.categories AS categories,
	s.status AS status,
	s.cover_media_id AS cover_media_id,
	s.rating_count AS rating_count,
	s.rating_avg AS rating_avg,
	s.rating_score AS rating_score,
	s.taste_avg AS taste_avg,
	s.value_avg AS value_avg,
	s.queue_avg AS queue_avg,
	s.hygiene_avg AS hygiene_avg,
	s.service_avg AS service_avg,
	s.review_event_count AS review_event_count,
	s.last_review_at AS last_review_at`

// CanteenStoreQueryRepository implements the canteen hall and store read models.
// It never touches review events for list queries; aggregates come from the
// persisted stats columns.
type CanteenStoreQueryRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewCanteenStoreQueryRepository(db *gorm.DB, cache ...*redis.Client) *CanteenStoreQueryRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CanteenStoreQueryRepository{db: db, cache: client}
}

func (r *CanteenStoreQueryRepository) ListHalls(ctx context.Context) ([]canteen.HallView, error) {
	var rows []canteenHallRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			h.id,
			h.campus_key,
			h.name,
			h.short_name,
			h.address,
			h.status,
			h.display_order,
			COALESCE(cnt.store_count, 0) AS store_count
		FROM canteen_halls AS h
		LEFT JOIN (
			SELECT hall_id, count(*) AS store_count
			FROM canteen_stores
			WHERE status = ?
			GROUP BY hall_id
		) AS cnt ON cnt.hall_id = h.id
		WHERE h.status <> ?
		ORDER BY h.campus_key, h.display_order, h.id
	`, string(canteen.StoreStatusActive), string(canteen.HallStatusHidden)).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	halls := make([]canteen.HallView, 0, len(rows))
	for _, row := range rows {
		halls = append(halls, canteen.HallView{
			ID:           row.ID,
			CampusKey:    row.CampusKey,
			Name:         row.Name,
			ShortName:    row.ShortName,
			Address:      row.Address,
			Status:       canteen.HallStatus(row.Status),
			DisplayOrder: row.DisplayOrder,
			StoreCount:   row.StoreCount,
		})
	}
	return halls, nil
}

func (r *CanteenStoreQueryRepository) ListStores(ctx context.Context, filter canteen.StoreListFilter) ([]canteen.StoreListItem, int64, error) {
	filter.Page, filter.PageSize = normalizeCanteenPage(filter.Page, filter.PageSize)
	query := r.storeListQuery(ctx, filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []canteenStoreListRow
	err := query.
		Select(canteenStoreListSelect).
		Order(canteenStoreOrder(filter.Sort)).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	items := make([]canteen.StoreListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, newCanteenStoreListItem(row))
	}
	return items, total, nil
}

type canteenRankingCache struct {
	Items []canteen.StoreListItem
	Total int64
}

func (r *CanteenStoreQueryRepository) ListRankings(ctx context.Context, filter canteen.RankingFilter) ([]canteen.StoreListItem, int64, error) {
	filter.Page, filter.PageSize = normalizeCanteenPage(filter.Page, filter.PageSize)
	cacheK := cacheKey("canteen", "rankings", "v2", filter.CampusKey, filter.HallID, filter.Sort, filter.Page, filter.PageSize)
	if cached, ok := cacheGetJSON[canteenRankingCache](ctx, r.cache, cacheK); ok {
		return cached.Items, cached.Total, nil
	}
	storeFilter := canteen.StoreListFilter{
		CampusKey: filter.CampusKey,
		HallID:    filter.HallID,
		Status:    canteen.StoreStatusActive,
		Sort:      filter.Sort,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
	}
	query := r.storeListQuery(ctx, storeFilter).Where("s.rating_count > 0")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []canteenStoreListRow
	err := query.
		Select(canteenStoreListSelect).
		Order(canteenStoreOrder(storeFilter.Sort)).
		Offset((storeFilter.Page - 1) * storeFilter.PageSize).
		Limit(storeFilter.PageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	items := make([]canteen.StoreListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, newCanteenStoreListItem(row))
	}
	if err := r.enrichRankingItems(ctx, items); err != nil {
		return nil, 0, err
	}
	cacheSetJSON(ctx, r.cache, cacheK, canteenRankingCache{Items: items, Total: total})
	return items, total, nil
}

// enrichRankingItems adds the explanation layer: sample-size confidence, the
// last 30 days' most frequent tags and the dish/photo coverage.
func (r *CanteenStoreQueryRepository) enrichRankingItems(ctx context.Context, items []canteen.StoreListItem) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int, 0, len(items))
	for i := range items {
		items[i].Confidence = canteen.RatingConfidence(items[i].RatingCount)
		ids = append(ids, items[i].ID)
	}

	type tagRow struct {
		StoreID int            `gorm:"column:store_id"`
		Tags    pq.StringArray `gorm:"column:tags;type:text[]"`
	}
	var tagRows []tagRow
	if err := r.db.WithContext(ctx).
		Table("canteen_review_events").
		Select("store_id, tags").
		Where("store_id IN ? AND status = ? AND created_at >= ?", ids, string(canteen.ReviewStatusActive), time.Now().AddDate(0, 0, -30)).
		Where("tags IS NOT NULL AND cardinality(tags) > 0").
		Scan(&tagRows).Error; err != nil {
		return err
	}
	tagCounts := make(map[int]map[string]int, len(ids))
	for _, row := range tagRows {
		counts := tagCounts[row.StoreID]
		if counts == nil {
			counts = map[string]int{}
			tagCounts[row.StoreID] = counts
		}
		for _, tag := range row.Tags {
			counts[tag]++
		}
	}
	for i := range items {
		counts := tagCounts[items[i].ID]
		type pair struct {
			tag   string
			count int
		}
		pairs := make([]pair, 0, len(counts))
		for tag, count := range counts {
			pairs = append(pairs, pair{tag: tag, count: count})
		}
		sort.Slice(pairs, func(a, b int) bool {
			if pairs[a].count != pairs[b].count {
				return pairs[a].count > pairs[b].count
			}
			return pairs[a].tag < pairs[b].tag
		})
		limit := len(pairs)
		if limit > 3 {
			limit = 3
		}
		for _, item := range pairs[:limit] {
			items[i].SummaryTags = append(items[i].SummaryTags, item.tag)
		}
	}

	type countRow struct {
		StoreID int   `gorm:"column:store_id"`
		Count   int64 `gorm:"column:count"`
	}
	var dishRows []countRow
	if err := r.db.WithContext(ctx).
		Table("canteen_dishes").
		Select("store_id, COUNT(*) AS count").
		Where("store_id IN ? AND status = ?", ids, string(canteen.DishStatusActive)).
		Group("store_id").
		Scan(&dishRows).Error; err != nil {
		return err
	}
	dishCounts := make(map[int]int, len(dishRows))
	for _, row := range dishRows {
		dishCounts[row.StoreID] = int(row.Count)
	}

	var photoRows []countRow
	if err := r.db.WithContext(ctx).
		Table("canteen_dish_media AS dm").
		Joins("JOIN canteen_dishes AS d ON d.id = dm.dish_id").
		Joins("JOIN canteen_media AS m ON m.id = dm.media_id").
		Select("d.store_id AS store_id, COUNT(DISTINCT m.id) AS count").
		Where("d.store_id IN ? AND m.status = ?", ids, string(canteen.MediaStatusApproved)).
		Group("d.store_id").
		Scan(&photoRows).Error; err != nil {
		return err
	}
	photoCounts := make(map[int]int, len(photoRows))
	for _, row := range photoRows {
		photoCounts[row.StoreID] = int(row.Count)
	}

	for i := range items {
		items[i].DishCount = dishCounts[items[i].ID]
		items[i].DishPhotoCount = photoCounts[items[i].ID]
	}
	return nil
}

func (r *CanteenStoreQueryRepository) GetStoreDetail(ctx context.Context, id int, viewerID int) (*canteen.StoreDetail, error) {
	cacheK := cacheKey("canteen", "store", id, "detail", "v2", viewerID)
	if cached, ok := cacheGetJSON[canteen.StoreDetail](ctx, r.cache, cacheK); ok {
		return cached, nil
	}
	var rows []canteenStoreListRow
	err := r.db.WithContext(ctx).
		Table("canteen_stores AS s").
		Joins("JOIN canteen_halls AS h ON h.id = s.hall_id").
		Where("s.id = ?", id).
		Select(canteenStoreListSelect).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	detail := &canteen.StoreDetail{
		StoreListItem: newCanteenStoreListItem(row),
		Dimensions: canteen.DimensionAverages{
			Taste:   row.TasteAvg,
			Value:   row.ValueAvg,
			Queue:   row.QueueAvg,
			Hygiene: row.HygieneAvg,
			Service: row.ServiceAvg,
		},
		Dishes: []canteen.DishView{},
	}
	store, err := NewCanteenStoreRepository(r.db).Get(ctx, id)
	if err != nil {
		return nil, err
	}
	detail.Store = store

	var dishRows []canteenDishRow
	if err := r.db.WithContext(ctx).
		Table("canteen_dishes AS d").
		Joins("JOIN canteen_stores AS s ON s.id = d.store_id").
		Where("d.store_id = ? AND d.status = ?", id, string(canteen.DishStatusActive)).
		Select(canteenDishSelect).
		Order("d.rating_score DESC, d.rating_count DESC, d.name ASC, d.id ASC").
		Scan(&dishRows).Error; err != nil {
		return nil, err
	}
	for _, dishRow := range dishRows {
		detail.Dishes = append(detail.Dishes, newCanteenDishView(dishRow))
	}

	if viewerID > 0 {
		var summary CanteenUserRatingSummaryEntity
		err := r.db.WithContext(ctx).
			Where("store_id = ? AND user_id = ?", id, viewerID).
			Take(&summary).Error
		switch {
		case err == nil:
			detail.MySummary = newCanteenUserSummaryDomain(&summary)
		case errors.Is(err, gorm.ErrRecordNotFound):
		default:
			return nil, err
		}

		var reviews []canteenReviewRow
		err = r.db.WithContext(ctx).
			Table("canteen_review_events AS e").
			Joins("JOIN canteen_stores AS s ON s.id = e.store_id").
			Where("e.store_id = ? AND e.user_id = ? AND e.status <> ?", id, viewerID, string(canteen.ReviewStatusDeleted)).
			Select(`
				e.id AS id,
				e.store_id AS store_id,
				s.name AS store_name,
				e.user_id AS user_id,
				e.visited_at AS visited_at,
				e.taste AS taste,
				e.value AS value,
				e.queue AS queue,
				e.hygiene AS hygiene,
				e.service AS service,
				e.overall_score AS overall_score,
				e.content AS content,
				e.tags AS tags,
				e.status AS status,
				e.like_count AS like_count,
				e.dislike_count AS dislike_count,
				e.created_at AS created_at,
				e.updated_at AS updated_at`).
			Order("e.updated_at DESC, e.id DESC").
			Limit(1).
			Scan(&reviews).Error
		if err != nil {
			return nil, err
		}
		if len(reviews) > 0 {
			views := []canteen.ReviewView{newCanteenReviewView(reviews[0])}
			if err := r.attachCanteenMedia(ctx, views); err != nil {
				return nil, err
			}
			detail.MyLatestReview = &views[0]
		}
	}
	cacheSetJSON(ctx, r.cache, cacheK, detail)
	return detail, nil
}

func (r *CanteenStoreQueryRepository) GetHome(ctx context.Context) (*canteen.HomeView, error) {
	cacheK := cacheKey("canteen", "home", "v2")
	if cached, ok := cacheGetJSON[canteen.HomeView](ctx, r.cache, cacheK); ok {
		return cached, nil
	}
	halls, err := r.ListHalls(ctx)
	if err != nil {
		return nil, err
	}
	topStores, _, err := r.ListRankings(ctx, canteen.RankingFilter{Sort: "rating_score", Page: 1, PageSize: 6})
	if err != nil {
		return nil, err
	}
	newStores, _, err := r.ListStores(ctx, canteen.StoreListFilter{
		Status:   canteen.StoreStatusActive,
		Sort:     "latest",
		Page:     1,
		PageSize: 6,
	})
	if err != nil {
		return nil, err
	}
	latestReviews, err := r.latestReviews(ctx, 5)
	if err != nil {
		return nil, err
	}

	var storeCount, reviewCount, dishCount int64
	if err := r.db.WithContext(ctx).Model(&CanteenStoreEntity{}).
		Where("status = ?", string(canteen.StoreStatusActive)).
		Count(&storeCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&CanteenReviewEventEntity{}).
		Where("status = ?", string(canteen.ReviewStatusActive)).
		Count(&reviewCount).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&CanteenDishEntity{}).
		Where("status = ?", string(canteen.DishStatusActive)).
		Count(&dishCount).Error; err != nil {
		return nil, err
	}
	stats := canteen.HomeStats{
		StoreCount:  int(storeCount),
		ReviewCount: int(reviewCount),
		DishCount:   int(dishCount),
	}

	hotDishes, err := r.loadHotDishes(ctx, 6)
	if err != nil {
		return nil, err
	}
	feeds := make([]canteen.HomeFeedItem, 0, 5)
	var hero *canteen.HomeHero
	if len(topStores) > 0 {
		first := topStores[0]
		imageURL := ""
		if first.CoverMediaID != nil {
			imageURL = fmt.Sprintf("/api/canteen/media/%d", *first.CoverMediaID)
		}
		hero = &canteen.HomeHero{
			Type:        "recommended_store",
			StoreID:     first.ID,
			StoreName:   first.Name,
			HallName:    first.HallName,
			ImageURL:    imageURL,
			Reason:      "综合评分领先，适合今天去试试",
			RatingScore: first.RatingScore,
			RatingCount: first.RatingCount,
		}
		for _, dish := range hotDishes {
			if dish.StoreID == first.ID && dish.CoverURL != "" {
				hero.DishID = dish.ID
				hero.DishName = dish.Name
				hero.ImageURL = dish.CoverURL
				hero.Reason = "招牌菜品实拍最多，评分稳定"
				break
			}
		}
	}
	if len(topStores) > 1 {
		next := topStores[1]
		feeds = append(feeds, canteen.HomeFeedItem{
			Type:        "recommended_store",
			StoreID:     next.ID,
			StoreName:   next.Name,
			HallName:    next.HallName,
			Reason:      "综合评分靠前，样本较充足",
			RatingScore: next.RatingScore,
			RatingCount: next.RatingCount,
			Tags:        next.SummaryTags,
		})
	}
	if len(newStores) > 0 {
		newest := newStores[0]
		feeds = append(feeds, canteen.HomeFeedItem{
			Type:        "new_store",
			StoreID:     newest.ID,
			StoreName:   newest.Name,
			HallName:    newest.HallName,
			Reason:      "最近收录的新店，等你打卡",
			RatingScore: newest.RatingScore,
			RatingCount: newest.RatingCount,
		})
	}
	if trending, err := r.trendingStore(ctx); err != nil {
		return nil, err
	} else if trending != nil {
		feeds = append(feeds, *trending)
	}
	for _, store := range topStores {
		if store.RatingCount >= 5 {
			feeds = append(feeds, canteen.HomeFeedItem{
				Type:        "stable_choice",
				StoreID:     store.ID,
				StoreName:   store.Name,
				HallName:    store.HallName,
				Reason:      "评价人数充足，表现稳定",
				RatingScore: store.RatingScore,
				RatingCount: store.RatingCount,
				Tags:        store.SummaryTags,
			})
			break
		}
	}
	if photo, err := r.recentPhotoFeed(ctx); err != nil {
		return nil, err
	} else if photo != nil {
		feeds = append(feeds, *photo)
	}

	home := &canteen.HomeView{
		Halls:         halls,
		TopStores:     topStores,
		NewStores:     newStores,
		LatestReviews: latestReviews,
		HotDishes:     hotDishes,
		Hero:          hero,
		Feeds:         feeds,
		Stats:         stats,
	}
	cacheSetJSON(ctx, r.cache, cacheK, home)
	return home, nil
}

func (r *CanteenStoreQueryRepository) loadHotDishes(ctx context.Context, limit int) ([]canteen.DishView, error) {
	var rows []canteenDishRow
	if err := r.db.WithContext(ctx).
		Table("canteen_dishes AS d").
		Joins("JOIN canteen_stores AS s ON s.id = d.store_id").
		Where("d.status = ? AND s.status = ? AND d.rating_count > 0", string(canteen.DishStatusActive), string(canteen.StoreStatusActive)).
		Select(canteenDishSelect).
		Order("d.rating_score DESC, d.rating_count DESC, d.id ASC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	dishes := make([]canteen.DishView, 0, len(rows))
	for _, row := range rows {
		dishes = append(dishes, newCanteenDishView(row))
	}
	return dishes, nil
}

func (r *CanteenStoreQueryRepository) trendingStore(ctx context.Context) (*canteen.HomeFeedItem, error) {
	type trendRow struct {
		StoreID   int    `gorm:"column:store_id"`
		StoreName string `gorm:"column:store_name"`
		HallName  string `gorm:"column:hall_name"`
		Count     int    `gorm:"column:count"`
	}
	var rows []trendRow
	if err := r.db.WithContext(ctx).
		Table("canteen_review_events AS e").
		Joins("JOIN canteen_stores AS s ON s.id = e.store_id").
		Joins("JOIN canteen_halls AS h ON h.id = s.hall_id").
		Where("e.status = ? AND s.status = ? AND e.created_at >= ?", string(canteen.ReviewStatusActive), string(canteen.StoreStatusActive), time.Now().AddDate(0, 0, -7)).
		Select("s.id AS store_id, s.name AS store_name, h.name AS hall_name, COUNT(*) AS count").
		Group("s.id, s.name, h.name").
		Order("count DESC, s.id ASC").
		Limit(1).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	return &canteen.HomeFeedItem{
		Type:      "trending",
		StoreID:   row.StoreID,
		StoreName: row.StoreName,
		HallName:  row.HallName,
		Reason:    "近 7 天到店评价最多",
	}, nil
}

func (r *CanteenStoreQueryRepository) recentPhotoFeed(ctx context.Context) (*canteen.HomeFeedItem, error) {
	type photoRow struct {
		StoreID   int    `gorm:"column:store_id"`
		StoreName string `gorm:"column:store_name"`
		HallName  string `gorm:"column:hall_name"`
		MediaID   int    `gorm:"column:media_id"`
	}
	var rows []photoRow
	if err := r.db.WithContext(ctx).
		Table("canteen_review_media AS rm").
		Joins("JOIN canteen_media AS m ON m.id = rm.media_id").
		Joins("JOIN canteen_review_events AS e ON e.id = rm.review_event_id").
		Joins("JOIN canteen_stores AS s ON s.id = e.store_id").
		Joins("JOIN canteen_halls AS h ON h.id = s.hall_id").
		Where("m.status = ? AND e.status = ? AND s.status = ?", string(canteen.MediaStatusApproved), string(canteen.ReviewStatusActive), string(canteen.StoreStatusActive)).
		Select("s.id AS store_id, s.name AS store_name, h.name AS hall_name, m.id AS media_id").
		Order("m.created_at DESC, m.id DESC").
		Limit(1).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	return &canteen.HomeFeedItem{
		Type:      "recent_photo",
		StoreID:   row.StoreID,
		StoreName: row.StoreName,
		HallName:  row.HallName,
		Reason:    "同学刚上传的实拍",
		ImageURL:  fmt.Sprintf("/api/canteen/media/%d", row.MediaID),
	}, nil
}

func (r *CanteenStoreQueryRepository) latestReviews(ctx context.Context, limit int) ([]canteen.ReviewView, error) {
	var rows []canteenReviewRow
	err := r.db.WithContext(ctx).
		Table("canteen_review_events AS e").
		Joins("JOIN canteen_stores AS s ON s.id = e.store_id").
		Where("e.status = ? AND s.status = ?", string(canteen.ReviewStatusActive), string(canteen.StoreStatusActive)).
		Select(`
			e.id AS id,
			e.store_id AS store_id,
			s.name AS store_name,
			e.user_id AS user_id,
			e.visited_at AS visited_at,
			e.taste AS taste,
			e.value AS value,
			e.queue AS queue,
			e.hygiene AS hygiene,
			e.service AS service,
			e.overall_score AS overall_score,
			e.content AS content,
			e.tags AS tags,
			e.status AS status,
			e.like_count AS like_count,
			e.dislike_count AS dislike_count,
			e.created_at AS created_at,
			e.updated_at AS updated_at`).
		Order("e.created_at DESC, e.id DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	reviews := make([]canteen.ReviewView, 0, len(rows))
	for _, row := range rows {
		reviews = append(reviews, newCanteenReviewView(row))
	}
	if err := r.attachCanteenMedia(ctx, reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *CanteenStoreQueryRepository) storeListQuery(ctx context.Context, filter canteen.StoreListFilter) *gorm.DB {
	query := r.db.WithContext(ctx).
		Table("canteen_stores AS s").
		Joins("JOIN canteen_halls AS h ON h.id = s.hall_id")

	status := filter.Status
	if status == "" {
		status = canteen.StoreStatusActive
	}
	query = query.Where("s.status = ?", string(status))
	if filter.CampusKey != "" {
		query = query.Where("h.campus_key = ?", filter.CampusKey)
	}
	if filter.HallID > 0 {
		query = query.Where("s.hall_id = ?", filter.HallID)
	}
	if filter.Category != "" {
		query = query.Where("? = ANY(s.categories)", filter.Category)
	}
	if term := canteen.NormalizeName(filter.Q); term != "" {
		query = query.Where("s.normalized_name LIKE ?", "%"+term+"%")
	}
	return query
}

func normalizeCanteenPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func canteenStoreOrder(sort string) string {
	switch sort {
	case "rating_count":
		return "s.rating_count DESC, s.rating_score DESC, s.id DESC"
	case "latest":
		return "s.created_at DESC, s.id DESC"
	case "name":
		return "s.name ASC, s.id ASC"
	default:
		return "s.rating_score DESC, s.rating_count DESC, s.id DESC"
	}
}

const canteenReviewSelect = `
	e.id AS id,
	e.store_id AS store_id,
	s.name AS store_name,
	e.user_id AS user_id,
	e.visited_at AS visited_at,
	e.taste AS taste,
	e.value AS value,
	e.queue AS queue,
	e.hygiene AS hygiene,
	e.service AS service,
	e.overall_score AS overall_score,
	e.content AS content,
	e.tags AS tags,
	e.status AS status,
	e.like_count AS like_count,
	e.dislike_count AS dislike_count,
	e.created_at AS created_at,
	e.updated_at AS updated_at`

func (r *CanteenStoreQueryRepository) ListReviews(ctx context.Context, filter canteen.ReviewListFilter) ([]canteen.ReviewView, int64, error) {
	filter.Page, filter.PageSize = normalizeCanteenPage(filter.Page, filter.PageSize)
	query := r.db.WithContext(ctx).
		Table("canteen_review_events AS e").
		Joins("JOIN canteen_stores AS s ON s.id = e.store_id")
	if filter.StoreID > 0 {
		query = query.Where("e.store_id = ?", filter.StoreID)
	}
	if filter.UserID > 0 {
		query = query.Where("e.user_id = ?", filter.UserID)
	}
	status := filter.Status
	if status == "" {
		status = canteen.ReviewStatusActive
	}
	query = query.Where("e.status = ?", string(status))
	if filter.WithMedia {
		query = query.Where("EXISTS (SELECT 1 FROM canteen_review_media m JOIN canteen_media cm ON cm.id = m.media_id WHERE m.review_event_id = e.id AND cm.status = ?)", string(canteen.MediaStatusApproved))
	}
	if filter.MinScore != nil {
		query = query.Where("e.overall_score >= ?", *filter.MinScore)
	}
	if filter.MaxScore != nil {
		query = query.Where("e.overall_score <= ?", *filter.MaxScore)
	}
	// The default store feed shows each user's latest visit. Applying the
	// filters after this subquery keeps e.g. "high score" honest: a user whose
	// latest visit was low does not reappear through an older high visit.
	if filter.StoreID > 0 && !filter.History {
		query = query.Where(`e.id = (
			SELECT e2.id FROM canteen_review_events e2
			WHERE e2.store_id = e.store_id AND e2.user_id = e.user_id AND e2.status = ?
			ORDER BY e2.created_at DESC, e2.id DESC
			LIMIT 1
		)`, string(status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if filter.Cursor != nil {
		cursor := *filter.Cursor
		switch filter.Order {
		case "best":
			query = query.Where(
				"(e.overall_score < ? OR (e.overall_score = ? AND (e.like_count < ? OR (e.like_count = ? AND e.id < ?))))",
				cursor.OverallScore, cursor.OverallScore, cursor.LikeCount, cursor.LikeCount, cursor.ID,
			)
		case "worst":
			query = query.Where(
				"(e.overall_score > ? OR (e.overall_score = ? AND e.id < ?))",
				cursor.OverallScore, cursor.OverallScore, cursor.ID,
			)
		default:
			query = query.Where(
				"(e.created_at < ? OR (e.created_at = ? AND e.id < ?))",
				cursor.CreatedAt, cursor.CreatedAt, cursor.ID,
			)
		}
	}
	pageQuery := query
	if filter.Cursor == nil {
		pageQuery = pageQuery.Offset((filter.Page - 1) * filter.PageSize)
	}
	var rows []canteenReviewRow
	if err := pageQuery.
		Select(canteenReviewSelect).
		Order(canteenReviewOrder(filter.Order)).
		Limit(filter.PageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	views := make([]canteen.ReviewView, 0, len(rows))
	for _, row := range rows {
		views = append(views, newCanteenReviewView(row))
	}
	if err := r.attachCanteenMedia(ctx, views); err != nil {
		return nil, 0, err
	}
	if err := r.attachMyVotes(ctx, views, filter.ViewerID); err != nil {
		return nil, 0, err
	}
	return views, total, nil
}

func (r *CanteenStoreQueryRepository) GetReview(ctx context.Context, id int) (*canteen.ReviewView, error) {
	var rows []canteenReviewRow
	if err := r.db.WithContext(ctx).
		Table("canteen_review_events AS e").
		Joins("JOIN canteen_stores AS s ON s.id = e.store_id").
		Where("e.id = ?", id).
		Select(canteenReviewSelect).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	views := []canteen.ReviewView{newCanteenReviewView(rows[0])}
	if err := r.attachCanteenMedia(ctx, views); err != nil {
		return nil, err
	}
	return &views[0], nil
}

func (r *CanteenStoreQueryRepository) ListUserReviews(ctx context.Context, userID int, page, pageSize int) ([]canteen.ReviewView, int64, error) {
	page, pageSize = normalizeCanteenPage(page, pageSize)
	query := r.db.WithContext(ctx).
		Table("canteen_review_events AS e").
		Joins("JOIN canteen_stores AS s ON s.id = e.store_id").
		Where("e.user_id = ? AND e.status <> ?", userID, string(canteen.ReviewStatusDeleted))
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []canteenReviewRow
	if err := query.
		Select(canteenReviewSelect).
		Order("e.updated_at DESC, e.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	views := make([]canteen.ReviewView, 0, len(rows))
	for _, row := range rows {
		views = append(views, newCanteenReviewView(row))
	}
	if err := r.attachCanteenMedia(ctx, views); err != nil {
		return nil, 0, err
	}
	if err := r.attachMyVotes(ctx, views, userID); err != nil {
		return nil, 0, err
	}
	return views, total, nil
}

func (r *CanteenStoreQueryRepository) GetUserSummary(ctx context.Context, storeID, userID int) (*canteen.UserRatingSummary, error) {
	var entity CanteenUserRatingSummaryEntity
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND user_id = ?", storeID, userID).
		Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenUserSummaryDomain(&entity), nil
}

func (r *CanteenStoreQueryRepository) CountSince(ctx context.Context, userID int, since time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CanteenReviewEventEntity{}).
		Where("user_id = ? AND created_at >= ? AND status <> ?", userID, since, string(canteen.ReviewStatusDeleted)).
		Count(&count).Error
	return int(count), err
}

func (r *CanteenStoreQueryRepository) ListRecentContent(ctx context.Context, userID int, since time.Time, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 5
	}
	var contents []string
	err := r.db.WithContext(ctx).Model(&CanteenReviewEventEntity{}).
		Where("user_id = ? AND created_at >= ? AND status <> ? AND content <> ''", userID, since, string(canteen.ReviewStatusDeleted)).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Pluck("content", &contents).Error
	return contents, err
}

func (r *CanteenStoreQueryRepository) attachMyVotes(ctx context.Context, views []canteen.ReviewView, viewerID int) error {
	if viewerID <= 0 || len(views) == 0 {
		return nil
	}
	ids := make([]int, 0, len(views))
	for _, view := range views {
		ids = append(ids, view.ID)
	}
	var votes []CanteenReviewVoteEntity
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND review_event_id IN ?", viewerID, ids).
		Find(&votes).Error; err != nil {
		return err
	}
	byReview := make(map[int]int, len(votes))
	for _, vote := range votes {
		byReview[vote.ReviewEventID] = vote.VoteType
	}
	for i := range views {
		views[i].MyVote = byReview[views[i].ID]
	}
	return nil
}

func canteenReviewOrder(order string) string {
	switch order {
	case "best":
		return "e.overall_score DESC, e.like_count DESC, e.id DESC"
	case "worst":
		return "e.overall_score ASC, e.id DESC"
	default:
		return "e.created_at DESC, e.id DESC"
	}
}

type canteenReviewMediaRow struct {
	ReviewEventID int `gorm:"column:review_event_id"`
	MediaID       int `gorm:"column:media_id"`
	Width         int `gorm:"column:width"`
	Height        int `gorm:"column:height"`
}

// attachCanteenMedia loads approved images for a page of reviews and builds the
// public media URL. Pending images never appear on review views.
func (r *CanteenStoreQueryRepository) attachCanteenMedia(ctx context.Context, views []canteen.ReviewView) error {
	if len(views) == 0 {
		return nil
	}
	ids := make([]int, 0, len(views))
	for _, view := range views {
		ids = append(ids, view.ID)
	}
	var rows []canteenReviewMediaRow
	if err := r.db.WithContext(ctx).
		Table("canteen_review_media AS rm").
		Joins("JOIN canteen_media AS m ON m.id = rm.media_id").
		Where("rm.review_event_id IN ? AND m.status = ?", ids, string(canteen.MediaStatusApproved)).
		Order("rm.sort_order ASC, m.id ASC").
		Select("rm.review_event_id, rm.media_id, m.width, m.height").
		Scan(&rows).Error; err != nil {
		return err
	}
	byReview := make(map[int][]canteen.MediaView, len(views))
	for _, row := range rows {
		byReview[row.ReviewEventID] = append(byReview[row.ReviewEventID], canteen.MediaView{
			ID:     row.MediaID,
			Width:  row.Width,
			Height: row.Height,
			URL:    fmt.Sprintf("/api/canteen/media/%d", row.MediaID),
		})
	}
	for i := range views {
		views[i].Media = byReview[views[i].ID]
	}
	return nil
}

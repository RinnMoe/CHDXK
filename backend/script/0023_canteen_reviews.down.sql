DROP TABLE IF EXISTS canteen_reports;
DROP TABLE IF EXISTS canteen_review_media;
DROP TABLE IF EXISTS canteen_dish_rating_summaries;
DROP TABLE IF EXISTS canteen_dish_review_events;
DROP TABLE IF EXISTS canteen_review_revisions;
DROP TABLE IF EXISTS canteen_review_votes;
DROP TABLE IF EXISTS canteen_user_rating_summaries;
DROP TABLE IF EXISTS canteen_review_dishes;
DROP TABLE IF EXISTS canteen_review_events;

ALTER TABLE canteen_stores DROP CONSTRAINT IF EXISTS fk_canteen_stores_cover;

DROP TABLE IF EXISTS canteen_media;

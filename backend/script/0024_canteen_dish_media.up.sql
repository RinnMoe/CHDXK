-- Dish photos reuse canteen_media; this table links approved/pending media to
-- a dish. Public galleries only expose media whose own status is approved.
CREATE TABLE IF NOT EXISTS canteen_dish_media
(
    id         SERIAL PRIMARY KEY,
    dish_id    INTEGER     NOT NULL,
    media_id   INTEGER     NOT NULL,
    user_id    INTEGER     NOT NULL,
    sort_order INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_canteen_dish_media_dish
        FOREIGN KEY (dish_id) REFERENCES canteen_dishes (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_canteen_dish_media_media
        FOREIGN KEY (media_id) REFERENCES canteen_media (id)
            ON DELETE CASCADE,
    CONSTRAINT uniq_canteen_dish_media UNIQUE (dish_id, media_id)
);

CREATE INDEX IF NOT EXISTS idx_canteen_dish_media_dish
    ON canteen_dish_media (dish_id);

CREATE INDEX IF NOT EXISTS idx_canteen_dish_media_media
    ON canteen_dish_media (media_id);

CREATE INDEX IF NOT EXISTS idx_canteen_dish_media_user
    ON canteen_dish_media (user_id);

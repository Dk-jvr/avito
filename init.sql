DROP TABLE IF EXISTS Banners;

CREATE TABLE IF NOT EXISTS Banners(
    banner_id SERIAL PRIMARY KEY,
    tag_ids INT[],
    feature_id INTEGER,
    content TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    is_active BOOLEAN
);

INSERT INTO Banners(tag_ids, feature_id, content, created_at, updated_at, is_active) VALUES (ARRAY[1, 2, 3], 1, 'first_banner', NOW(), NOW(), true);
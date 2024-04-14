DROP TABLE IF EXISTS Banners;
DROP TABLE IF EXISTS OldBanners;

CREATE TABLE IF NOT EXISTS Banners(
    banner_id SERIAL PRIMARY KEY,
    tag_ids INT[],
    feature_id INTEGER,
    content TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    is_active BOOLEAN,
    version INT DEFAULT 1
);

INSERT INTO Banners(tag_ids, feature_id, content, created_at, updated_at, is_active) VALUES (ARRAY[1, 2, 3], 1, 'test banner', NOW(), NOW(), true);

CREATE TABLE IF NOT EXISTS OldBanners(
    banner_id INT,
    tag_ids INT[],
    feature_id INTEGER,
    content TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    is_active BOOLEAN,
    version INT
);

CREATE INDEX feature_id_index ON Banners USING HASH (feature_id);
CREATE INDEX tag_ids_index ON Banners USING GIN (tag_ids);

CREATE INDEX feature_id_index ON OldBanners USING HASH (feature_id);
CREATE INDEX tag_ids_index ON OldBanners USING GIN (tag_ids);


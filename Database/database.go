package Database

import (
	"avito/Components"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
	"strconv"
	"strings"
	"sync"
)

var (
	db    *sql.DB
	mutex sync.Mutex
)

func InitDataBase() *sql.DB {
	var err error
	const connectionString = `user=user dbname=pg_banners password=password host=db port=5432 sslmode=disable`
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	/*const queryInitDatabase = `
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
			INSERT INTO Banners(tag_ids, feature_id, content, created_at, updated_at, is_active) VALUES (ARRAY[1, 2, 3], 1, 'test banner', NOW(), NOW(), true);
		`
	_, err = db.Exec(queryInitDatabase)
	if err != nil {
		log.Fatal(err)
	}*/
	return db
}

func GetUserBanner(tagId int64, featureId int64, lastVersion bool, token string) (Components.ShortBanner, error) {
	var querySelectBanner string
	banner := new(Components.ShortBanner)
	mutex.Lock()
	defer mutex.Unlock()
	if lastVersion {
		querySelectBanner = `SELECT banner_id, content FROM Banners
									WHERE $1 = ANY(tag_ids) AND feature_id = $2 AND updated_at >= (NOW() - INTERVAL '5 MINUTES')`
	} else {
		querySelectBanner = `SELECT banner_id, content FROM Banners
									WHERE $1 = ANY(tag_ids) AND feature_id = $2 AND updated_at <= (NOW() - INTERVAL '5 MINUTES')`
	}
	if token == "admin_token" {
		querySelectBanner += `	
									ORDER BY updated_at DESC
									LIMIT 1;`
	} else {
		querySelectBanner += ` AND is_active = TRUE
									ORDER BY updated_at DESC
									LIMIT 1;`
	}
	err := db.QueryRow(querySelectBanner, tagId, featureId).Scan(&banner.BannerId, &banner.Content)
	if err != nil {
		return Components.ShortBanner{}, err
	}
	return *banner, nil
}

func GetBanners(tagId, featureId, limit, offset int64, token string) ([]Components.Banner, error) {
	var (
		querySelectBanners = `SELECT banner_id, array_to_string(tag_ids, ','), feature_id, content,
       		TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS'), TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI:SS'), is_active FROM Banners `
		tagIdsStr string
		banners   []Components.Banner
		rows      *sql.Rows
		err       error
	)
	if tagId != 0 && featureId != 0 {
		querySelectBanners += `	WHERE $1 = ANY(tag_ids) AND feature_id = $2`
		if token != "admin_token" {
			querySelectBanners += ` AND is_active = TRUE`
		}
		if limit != 0 {
			querySelectBanners += `	ORDER BY updated_at DESC
							LIMIT $3 OFFSET $4;`
			rows, err = db.Query(querySelectBanners, tagId, featureId, limit, offset)
		} else {
			querySelectBanners += `	ORDER BY updated_at DESC
							OFFSET $3;`
			rows, err = db.Query(querySelectBanners, tagId, featureId, offset)
		}

	} else if tagId != 0 {
		querySelectBanners += `	WHERE $1 = ANY(tag_ids)`
		if token != "admin_token" {
			querySelectBanners += ` AND is_active = TRUE`
		}
		if limit != 0 {
			querySelectBanners += `	ORDER BY updated_at DESC
							LIMIT $2 OFFSET $3;`
			rows, err = db.Query(querySelectBanners, tagId, limit, offset)
		} else {
			querySelectBanners += `	ORDER BY updated_at DESC
							OFFSET $2;`
			rows, err = db.Query(querySelectBanners, tagId, offset)
		}
	} else if featureId != 0 {
		querySelectBanners += `	WHERE feature_id = $1`
		if token != "admin_token" {
			querySelectBanners += ` AND is_active = TRUE`
		}
		if limit != 0 {
			querySelectBanners += `	ORDER BY updated_at DESC
							LIMIT $2 OFFSET $3;`
			rows, err = db.Query(querySelectBanners, featureId, limit, offset)
		} else {
			querySelectBanners += `	ORDER BY updated_at DESC
							OFFSET $2;`
			rows, err = db.Query(querySelectBanners, featureId, offset)
		}
	} else {
		if token != "admin_token" {
			querySelectBanners += ` WHERE is_active = TRUE`
		}
		if limit != 0 {
			querySelectBanners += `	ORDER BY updated_at DESC
							LIMIT $1 OFFSET $2;`
			rows, err = db.Query(querySelectBanners, limit, offset)
		} else {
			querySelectBanners += `	ORDER BY updated_at DESC
							OFFSET $1;`
			rows, err = db.Query(querySelectBanners, offset)
		}
	}
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		banner := new(Components.Banner)

		err = rows.Scan(&banner.BannerId, &tagIdsStr, &banner.FeatureId,
			&banner.Content, &banner.CreatedAtString, &banner.UpdatedAtString, &banner.IsActive)
		tagIdsStr = strings.Trim(tagIdsStr, "{}")
		tagIdsSlice := strings.Split(tagIdsStr, ",")

		tagIds := make([]int, len(tagIdsSlice))
		for i, tagIdStr := range tagIdsSlice {
			tagIds[i], err = strconv.Atoi(tagIdStr)
			if err != nil {
				return nil, err
			}
		}
		banner.TagIds = tagIds
		banners = append(banners, *banner)
	}
	return banners, nil
}

func CreateBanner(banner Components.Banner) error {
	const queryCreateBanner = `INSERT INTO Banners(tag_ids, feature_id, content, created_at, updated_at, is_active) VALUES ($1, $2, $3, NOW(), NOW(), $4)`
	_, err := db.Exec(queryCreateBanner, pq.Array(banner.TagIds), banner.FeatureId, banner.Content, banner.IsActive)
	return err
}

func DeleteBanner(bannerId int64) error {
	var rowsAffected int64
	const queryDeleteBanner = `DELETE FROM Banners WHERE banner_id = $1`
	result, err := db.Exec(queryDeleteBanner, bannerId)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func UpdateBanner(bannerId int64, bannerData map[string]interface{}) error {
	var (
		queryInsertOldVersion = `INSERT INTO OldBanners(banner_id, tag_ids, feature_id, content, created_at, updated_at, is_active, version) 
													(SELECT banner_id, tag_ids, feature_id, content, created_at, updated_at, is_active, version FROM Banners WHERE banner_id = $1);
													`
		queryUpdateBannerData = ` UPDATE Banners SET `
		values                []interface{}
		rowsaffected          int64
		result                sql.Result
	)
	for key, value := range bannerData {
		if arrayValue, ok := value.([]interface{}); ok {
			arrayString := fmt.Sprintf("{%v}", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(arrayValue)), ","), "[]"))
			queryUpdateBannerData += fmt.Sprintf(`%s = $%d,`, key, len(values)+1)
			values = append(values, arrayString)
		} else {
			queryUpdateBannerData += fmt.Sprintf(`%s = $%d,`, key, len(values)+1)
			values = append(values, value)
		}
	}

	queryUpdateBannerData += fmt.Sprintf(` updated_at = NOW(), version = version + 1 WHERE banner_id = $%d`, len(values)+1)

	values = append(values, bannerId)
	_, err := db.Exec(queryInsertOldVersion, bannerId)
	if err != nil {
		return err
	}
	result, err = db.Exec(queryUpdateBannerData, values...)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if rowsaffected, err = result.RowsAffected(); rowsaffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func GetOldBanners(bannerId int64) ([]Components.Banner, error) {
	var (
		tagIdsStr string
		banners   []Components.Banner
	)
	const queryGetOldBanners = `SELECT banner_id, array_to_string(tag_ids, ','), feature_id, content,
       		TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS'), TO_CHAR(updated_at, 'YYYY-MM-DD HH24:MI:SS'), is_active, version FROM OldBanners WHERE banner_id = $1 ORDER BY version DESC LIMIT 3;`
	rows, err := db.Query(queryGetOldBanners, bannerId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		banner := new(Components.Banner)

		err = rows.Scan(&banner.BannerId, &tagIdsStr, &banner.FeatureId,
			&banner.Content, &banner.CreatedAtString, &banner.UpdatedAtString, &banner.IsActive, &banner.Version)
		tagIdsStr = strings.Trim(tagIdsStr, "{}")
		tagIdsSlice := strings.Split(tagIdsStr, ",")

		tagIds := make([]int, len(tagIdsSlice))
		for i, tagIdStr := range tagIdsSlice {
			tagIds[i], err = strconv.Atoi(tagIdStr)
			if err != nil {
				return nil, err
			}
		}
		banner.TagIds = tagIds
		banners = append(banners, *banner)
	}
	return banners, nil
}

func ReturnOldBanner(bannerId, version int64) error {
	const queryReturnOldBanner = `
			UPDATE Banners
			SET 
				tag_ids = (SELECT tag_ids FROM OldBanners WHERE banner_id = $1 AND version = $2),
				feature_id = (SELECT feature_id FROM OldBanners WHERE banner_id = $1 AND version = $2),
				content = (SELECT content FROM OldBanners WHERE banner_id = $1 AND version = $2),
				created_at = (SELECT created_at FROM OldBanners WHERE banner_id = $1 AND version = $2),
				version = (SELECT version FROM OldBanners WHERE banner_id = $1 AND version = $2),
				is_active = (SELECT is_active FROM OldBanners WHERE banner_id = $1 AND version = $2),
				updated_at = NOW()
			WHERE banner_id = $1
		`
	result, err := db.Exec(queryReturnOldBanner, bannerId, version)
	if err != nil {
		return err
	}
	if rowsaffected, _ := result.RowsAffected(); rowsaffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

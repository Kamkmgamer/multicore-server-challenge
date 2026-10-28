package main

// The list queries deliberately select the small page of article IDs before
// doing the per-article projections.  A list response is one database round
// trip: total count and page rows are returned together.  This matters at the
// workload's request rate because the old handler ran the page and count as
// separate pool operations, and the feed predicate could scan and sort all
// articles written by followed authors.

import "strings"

type listMode uint8

const (
	listAllMode listMode = iota
	listTagMode
	listAuthorMode
	listFavoriteMode
	listFeedMode
)

const articleListSelect = `SELECT a.id, a.slug, a.title, a.description, a.created_at, a.updated_at,
       ARRAY(SELECT (SELECT t.name FROM tags t WHERE t.id=at.tag_id) AS name
              FROM article_tags at WHERE at.article_id=a.id ORDER BY name) AS tag_list,
       fav.n AS favorites_count, fav.favorited,
       u.username, u.bio, u.image,
       EXISTS (SELECT 1 FROM follows fo WHERE fo.follower_id = $1 AND fo.followee_id = a.author_id) AS following
  FROM articles a
  JOIN users u ON u.id = a.author_id
  CROSS JOIN LATERAL (SELECT count(*) AS n, COALESCE(bool_or(f.user_id=$1),false) AS favorited
                     FROM favorites f WHERE f.article_id=a.id) fav`

const (
	listAllPage = articleListSelect + `
 WHERE a.id IN (SELECT p.id FROM articles p ORDER BY p.id DESC LIMIT $2 OFFSET $3)
 ORDER BY a.id DESC`
	listAllCount = `SELECT count(*) FROM (SELECT 1 FROM articles LIMIT 1000) s`

	listTagPage = articleListSelect + `
 WHERE a.id IN (SELECT at.article_id FROM article_tags at
                 WHERE at.tag_id = (SELECT t.id FROM tags t WHERE t.name = $4)
                 ORDER BY at.article_id DESC LIMIT $2 OFFSET $3)
 ORDER BY a.id DESC`
	listTagCount = `SELECT count(*) FROM (SELECT 1 FROM article_tags at
                 WHERE at.tag_id = (SELECT t.id FROM tags t WHERE t.name = $1)
                 LIMIT 1000) s`

	listAuthorPage = articleListSelect + `
 WHERE a.id IN (SELECT p.id FROM articles p
                 WHERE p.author_id = (SELECT au.id FROM users au WHERE au.username = $4)
                 ORDER BY p.id DESC LIMIT $2 OFFSET $3)
 ORDER BY a.id DESC`
	listAuthorCount = `SELECT count(*) FROM (SELECT 1 FROM articles a
                   WHERE a.author_id = (SELECT u.id FROM users u WHERE u.username = $1)
                   LIMIT 1000) s`

	listFavoritePage = articleListSelect + `
 WHERE a.id IN (SELECT f.article_id FROM favorites f
                 WHERE f.user_id = (SELECT fu.id FROM users fu WHERE fu.username = $4)
                 ORDER BY f.article_id DESC LIMIT $2 OFFSET $3)
 ORDER BY a.id DESC`
	listFavoriteCount = `SELECT count(*) FROM (SELECT 1 FROM favorites f
                    WHERE f.user_id = (SELECT fu.id FROM users fu WHERE fu.username = $1)
                    LIMIT 1000) s`

	// For each followed author, fetch only enough recent IDs to satisfy this
	// page, then merge those small sets.  The articles_author_idx index makes
	// each lateral lookup an ordered index walk instead of a broad filter/sort.
	listFeedPage = articleListSelect + `
 WHERE a.id IN (SELECT x.id
                 FROM follows fo
                 CROSS JOIN LATERAL (
                   SELECT a2.id FROM articles a2
                    WHERE a2.author_id = fo.followee_id
                    ORDER BY a2.id DESC
                    LIMIT (LEAST($2::bigint + $3::bigint,
                                9223372036854775807)::bigint)
                 ) x
                WHERE fo.follower_id = $1
                ORDER BY x.id DESC LIMIT $2 OFFSET $3)
 ORDER BY a.id DESC`
	listFeedCount = `SELECT count(*) FROM (SELECT 1 FROM articles a
                  WHERE a.author_id IN (SELECT fo.followee_id FROM follows fo
                                        WHERE fo.follower_id = $1)
                  LIMIT 1000) s`
)

// listQuery adds a capped total count and keeps one NULL-safe row when the
// offset is past the end.  The article id is zero in that sentinel row; all
// real article IDs are positive, so the handler can skip it without nullable
// scans or a second query.
func buildListQuery(mode listMode) string {
	var page, count string
	switch mode {
	case listTagMode:
		page, count = listTagPage, strings.ReplaceAll(listTagCount, "$1", "$4")
	case listAuthorMode:
		page, count = listAuthorPage, strings.ReplaceAll(listAuthorCount, "$1", "$4")
	case listFavoriteMode:
		page, count = listFavoritePage, strings.ReplaceAll(listFavoriteCount, "$1", "$4")
	case listFeedMode:
		page, count = listFeedPage, listFeedCount
	default:
		page, count = listAllPage, listAllCount
	}
	return `WITH total AS (` + count + `)
SELECT total.count,
       COALESCE(page.id, 0::bigint),
       COALESCE(page.slug, ''::text),
       COALESCE(page.title, ''::text),
       COALESCE(page.description, ''::text),
       COALESCE(page.created_at, 'epoch'::timestamptz),
       COALESCE(page.updated_at, 'epoch'::timestamptz),
       COALESCE(page.tag_list, '{}'::text[]),
       COALESCE(page.favorites_count, 0::bigint),
       COALESCE(page.favorited, false),
       COALESCE(page.username, ''::text),
       page.bio, page.image,
       COALESCE(page.following, false)
  FROM total
  LEFT JOIN LATERAL (` + page + `) page ON true
 ORDER BY page.id DESC`
}

var listQueries = func() map[listMode]string {
	queries := make(map[listMode]string, 5)
	for _, mode := range []listMode{listAllMode, listTagMode, listAuthorMode, listFavoriteMode, listFeedMode} {
		queries[mode] = buildListQuery(mode)
	}
	return queries
}()

func listQuery(mode listMode) string { return listQueries[mode] }

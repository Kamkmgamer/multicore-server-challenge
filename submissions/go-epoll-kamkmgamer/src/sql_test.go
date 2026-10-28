package main

import (
	"strings"
	"testing"
)

func TestListQueriesReturnCountAndPageInOneStatement(t *testing.T) {
	for _, mode := range []listMode{listAllMode, listTagMode, listAuthorMode, listFavoriteMode, listFeedMode} {
		q := listQuery(mode)
		if strings.Count(q, "WITH total AS (") != 1 {
			t.Fatalf("mode %d does not have one total CTE", mode)
		}
		if !strings.Contains(q, "LEFT JOIN LATERAL") {
			t.Fatalf("mode %d does not preserve the empty-page sentinel", mode)
		}
		if !strings.Contains(q, "COALESCE(page.id, 0::bigint)") {
			t.Fatalf("mode %d does not expose the empty-page sentinel", mode)
		}
	}
	for _, mode := range []listMode{listAllMode, listFeedMode} {
		if strings.Contains(listQuery(mode), "$4") {
			t.Fatalf("mode %d unexpectedly has a filter parameter", mode)
		}
	}
}

func TestListQueryUsesFeedTopKAndFilterParameter(t *testing.T) {
	feed := listQuery(listFeedMode)
	if !strings.Contains(feed, "CROSS JOIN LATERAL") || !strings.Contains(feed, "SELECT a2.id FROM articles a2") {
		t.Fatal("feed query does not select per-author recent IDs")
	}
	if !strings.Contains(feed, "ORDER BY x.id DESC LIMIT $2 OFFSET $3") {
		t.Fatal("feed query does not merge the bounded per-author ID sets")
	}

	for _, mode := range []listMode{listTagMode, listAuthorMode, listFavoriteMode} {
		q := listQuery(mode)
		if !strings.Contains(q, "$4") {
			t.Fatalf("mode %d does not bind its filter", mode)
		}
		if strings.Contains(q, "WHERE at.tag_id = (SELECT t.id FROM tags t WHERE t.name = $1)") ||
			strings.Contains(q, "fu.id WHERE fu.username = $1") {
			t.Fatalf("mode %d leaked the count query's filter placeholder", mode)
		}
	}
}

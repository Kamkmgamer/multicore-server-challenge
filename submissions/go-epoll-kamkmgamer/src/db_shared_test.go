package main

import "testing"

func sharedJob(stmt stmtID, viewer int64, limit, offset byte) *reactorJob {
	j := &reactorJob{}
	j.db.stmt = stmt
	j.db.viewer = viewer
	j.db.nparams = 3
	j.db.params[1] = []byte{limit}
	j.db.params[2] = []byte{offset}
	return j
}

func TestSameSharedRequest(t *testing.T) {
	tagsA := &reactorJob{}
	tagsA.db.stmt = stTags
	tagsB := &reactorJob{}
	tagsB.db.stmt = stTags
	if !sameSharedRequest(tagsA, tagsB) {
		t.Fatal("tag reads should share")
	}

	listA := sharedJob(stListAll, 0, 20, 40)
	listB := sharedJob(stListAll, 0, 20, 40)
	if !sameSharedRequest(listA, listB) {
		t.Fatal("identical anonymous list reads should share")
	}

	for _, stmt := range []stmtID{stProfile, stArticleBySlug, stComments} {
		a := sharedJob(stmt, 0, 0, 0)
		a.db.nparams = 2
		a.db.params[1] = []byte("same-slug")
		b := sharedJob(stmt, 0, 0, 0)
		b.db.nparams = 2
		b.db.params[1] = []byte("same-slug")
		if !sameSharedRequest(a, b) {
			t.Fatalf("anonymous statement %d reads should share", stmt)
		}
	}

	cases := []struct {
		name string
		b    *reactorJob
	}{
		{name: "viewer", b: sharedJob(stListAll, 7, 20, 40)},
		{name: "limit", b: sharedJob(stListAll, 0, 10, 40)},
		{name: "offset", b: sharedJob(stListAll, 0, 20, 80)},
		{name: "different statement", b: sharedJob(stListTag, 0, 20, 40)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if sameSharedRequest(listA, tc.b) {
				t.Fatalf("%s reads must not share", tc.name)
			}
		})
	}
}

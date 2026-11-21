package anilist

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/Hanivan/datami/types"
)

func bm(id, mal int, extra string) string {
	s := `{"id":` + itoa(id) + `,"idMal":` + itoa(mal) + `,"title":{"romaji":"t"}`
	if extra != "" {
		s += "," + extra
	}
	return s + "}"
}

func itoa(i int) string { return strconv.Itoa(i) }

func page(medias ...string) string {
	return `{"data":{"Page":{"media":[` + strings.Join(medias, ",") + `]}}}`
}

func TestGetByMalIDs_OneRequestKeyedByMalID(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return page(bm(102, 2, ""), bm(101, 1, "")) })
	got, err := c.GetByMalIDs(context.Background(), []string{"1", "2", "3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got["1"] == nil || got["2"] == nil {
		t.Fatalf("got %v", got)
	}
	if len(log.queries) != 1 || !strings.Contains(log.queries[0], "idMal_in") {
		t.Fatalf("queries %v", log.queries)
	}
	if log.vars[0]["type"] != "ANIME" || log.vars[0]["perPage"] != float64(3) {
		t.Fatalf("vars %v", log.vars[0])
	}
}

func TestGetByMalIDs_EmbedsPagesPerMedia(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string {
		return page(
			bm(101, 1, `"c1":`+conn(true, cEdge(1))+`,"c2":`+conn(false, cEdge(2))+`,"s1":`+conn(false, sEdge(1))),
			bm(102, 2, `"c1":`+conn(false, cEdge(3))+`,"c2":null`),
		)
	})
	got, err := c.GetByMalIDs(context.Background(), []string{"1", "2"}, types.MediaOpts{CharacterPages: 2, StaffPages: 3})
	if err != nil {
		t.Fatal(err)
	}
	q := log.queries[0]
	i := strings.Index(q, "media(idMal_in")
	if !strings.Contains(q[i:], "c2:") || !strings.Contains(q[i:], "s3:") {
		t.Fatalf("query %s", q)
	}
	if len(got["1"].Characters) != 2 || len(got["2"].Characters) != 1 || len(got["1"].Staff) != 1 {
		t.Fatalf("merge: %d %d %d", len(got["1"].Characters), len(got["2"].Characters), len(got["1"].Staff))
	}
}

func TestGetByMalIDs_FollowUpOnlyForMediaWithMore(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(n int) string {
		if n == 1 {
			return page(
				bm(101, 1, `"c1":`+conn(true, cEdge(1))+`,"c2":`+conn(true, cEdge(2))),
				bm(102, 2, `"c1":`+conn(false, cEdge(3))+`,"c2":`+conn(false)),
			)
		}
		return `{"data":{"Media":{"characters":` + conn(false, cEdge(9)) + `}}}`
	})
	got, err := c.GetByMalIDs(context.Background(), []string{"1", "2"}, types.MediaOpts{CharacterPages: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(log.queries) != 2 {
		t.Fatalf("requests %d", len(log.queries))
	}
	if log.vars[1]["id"] != float64(101) || log.vars[1]["page"] != float64(3) {
		t.Fatalf("follow-up vars %v", log.vars[1])
	}
	if len(got["1"].Characters) != 3 {
		t.Fatalf("chars %d", len(got["1"].Characters))
	}
}

func TestGetByMalIDs_Empty(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return page() })
	got, err := c.GetByMalIDs(context.Background(), nil)
	if err != nil || got == nil || len(got) != 0 || len(log.queries) != 0 {
		t.Fatalf("%v %v %d", got, err, len(log.queries))
	}
}

func TestGetByMalIDs_TooMany(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return page() })
	ids := make([]string, 51)
	for i := range ids {
		ids[i] = "1"
	}
	if _, err := c.GetByMalIDs(context.Background(), ids); !errors.Is(err, types.ErrInvalidInput) || len(log.queries) != 0 {
		t.Fatalf("%v %d", err, len(log.queries))
	}
	if _, err := c.GetByMalIDs(context.Background(), []string{"x"}); !errors.Is(err, types.ErrInvalidInput) {
		t.Fatal(err)
	}
}

func TestGetMangaByMalIDs_TypeManga_NoStaff(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return page(bm(101, 1, `"c1":`+conn(false, cEdge(1)))) })
	got, err := c.GetMangaByMalIDs(context.Background(), []string{"1"}, types.MediaOpts{CharacterPages: 2})
	if err != nil || got["1"] == nil {
		t.Fatal(err)
	}
	if log.vars[0]["type"] != "MANGA" || strings.Contains(log.queries[0], "s1:") || !strings.Contains(log.queries[0], "c2:") {
		t.Fatalf("%v %s", log.vars[0], log.queries[0])
	}
}

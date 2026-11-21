package anilist

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
	"golang.org/x/time/rate"
)

var aliasRe = regexp.MustCompile(`(a\d+): Media\(id: (\d+)\) \{ (characters|staff)\(page: (\d+), perPage: 25, sort: (ROLE|RELEVANCE)\)`)

type fuCall struct {
	id    int
	staff bool
	page  int
}

// followUpServer answers the first (batch or single) request with first, and
// every packed follow-up by asking answer for each alias. answer returns the
// connection JSON or "null" for a null Media; status != 0 fails that request.
type followUpServer struct {
	first   string
	errs    string // raw "errors" JSON added to follow-up replies, if set
	answer  func(fuCall) string
	status  func(n int) int
	queries []string
	rounds  [][]fuCall
}

func (s *followUpServer) client(t *testing.T) *Client {
	t.Helper()
	return s.clientWith(t, nil)
}

func (s *followUpServer) clientWith(t *testing.T, metrics func(string, string, time.Duration, error)) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var p struct {
			Query string `json:"query"`
		}
		json.Unmarshal(b, &p)
		s.queries = append(s.queries, p.Query)
		if s.status != nil {
			if code := s.status(len(s.queries)); code != 0 {
				w.WriteHeader(code)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if len(s.queries) == 1 {
			io.WriteString(w, s.first)
			return
		}
		var calls []fuCall
		var parts []string
		for _, m := range aliasRe.FindAllStringSubmatch(p.Query, -1) {
			id, _ := strconv.Atoi(m[2])
			pg, _ := strconv.Atoi(m[4])
			call := fuCall{id: id, staff: m[3] == "staff", page: pg}
			calls = append(calls, call)
			field := "characters"
			if call.staff {
				field = "staff"
			}
			v := s.answer(call)
			if v != "null" {
				v = `{"` + field + `":` + v + `}`
			}
			parts = append(parts, `"`+m[1]+`":`+v)
		}
		s.rounds = append(s.rounds, calls)
		reply := `{"data":{` + strings.Join(parts, ",") + `}`
		if s.errs != "" {
			reply += `,"errors":` + s.errs
		}
		io.WriteString(w, reply+`}`)
	}))
	t.Cleanup(srv.Close)
	return NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second, Limiter: rate.NewLimiter(rate.Inf, 1), Metrics: metrics})
}

// edgeID gives every (media, connection, page) edge a distinct id so order
// and ownership are checkable.
func edgeID(id int, staff bool, pg int) int {
	n := id*1000 + pg
	if staff {
		n += 500
	}
	return n
}

func embedded(id int, staff bool, pages int, next bool) string {
	var parts []string
	for pg := 1; pg <= pages; pg++ {
		alias, edge := "c", cEdge(edgeID(id, staff, pg))
		if staff {
			alias, edge = "s", sEdge(edgeID(id, staff, pg))
		}
		parts = append(parts, `"`+alias+itoa(pg)+`":`+conn(pg < pages || next, edge))
	}
	return strings.Join(parts, ",")
}

func followEdge(c fuCall, next bool) string {
	if c.staff {
		return conn(next, sEdge(edgeID(c.id, true, c.page)))
	}
	return conn(next, cEdge(edgeID(c.id, false, c.page)))
}

func checkPages(t *testing.T, label string, got []string, id int, staff bool, pages int) {
	t.Helper()
	if len(got) != pages {
		t.Fatalf("%s: %d edges, want %d", label, len(got), pages)
	}
	for i, g := range got {
		if want := itoa(edgeID(id, staff, i+1)); g != want {
			t.Errorf("%s: edge %d = %s, want %s", label, i, g, want)
		}
	}
}

func charIDs(m *types.Media) []string {
	var out []string
	for _, c := range m.Characters {
		out = append(out, c.ID)
	}
	return out
}

func staffIDs(m *types.Media) []string {
	var out []string
	for _, p := range m.Staff {
		out = append(out, p.ID)
	}
	return out
}

func TestFollowUps_PackedAcrossWorks(t *testing.T) {
	s := &followUpServer{
		first: page(
			bm(101, 1, embedded(101, false, 2, true)),
			bm(102, 2, embedded(102, false, 2, true)),
			bm(103, 3, embedded(103, false, 2, true)),
		),
		answer: func(c fuCall) string { return followEdge(c, c.page < 4) },
	}
	got, err := s.client(t).GetByMalIDs(context.Background(), []string{"1", "2", "3"}, types.MediaOpts{CharacterPages: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.rounds) != 2 || len(s.rounds[0]) != 3 || len(s.rounds[1]) != 3 {
		t.Fatalf("rounds %v", s.rounds)
	}
	for i, id := range []int{101, 102, 103} {
		checkPages(t, "media "+itoa(id), charIDs(got[itoa(i+1)]), id, false, 4)
	}
}

func TestFollowUps_SplitsAtAliasLimit(t *testing.T) {
	var medias []string
	ids := []string{}
	for i := 1; i <= 5; i++ {
		id := 100 + i
		medias = append(medias, bm(id, i, embedded(id, false, 2, true)+","+embedded(id, true, 3, true)))
		ids = append(ids, itoa(i))
	}
	s := &followUpServer{first: page(medias...), answer: func(c fuCall) string { return followEdge(c, false) }}
	got, err := s.client(t).GetByMalIDs(context.Background(), ids, types.MediaOpts{CharacterPages: 2, StaffPages: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.rounds) != 2 || len(s.rounds[0]) != followUpAliases || len(s.rounds[1]) != 2 {
		t.Fatalf("rounds %v", s.rounds)
	}
	for i := 1; i <= 5; i++ {
		checkPages(t, "chars", charIDs(got[itoa(i)]), 100+i, false, 3)
		checkPages(t, "staff", staffIDs(got[itoa(i)]), 100+i, true, 4)
	}
}

func TestFollowUps_DropOutIndependently(t *testing.T) {
	s := &followUpServer{
		first: page(bm(101, 1, embedded(101, false, 2, true)), bm(102, 2, embedded(102, false, 2, true))),
		answer: func(c fuCall) string {
			end := 3
			if c.id == 102 {
				end = 5
			}
			return followEdge(c, c.page < end)
		},
	}
	got, err := s.client(t).GetByMalIDs(context.Background(), []string{"1", "2"}, types.MediaOpts{CharacterPages: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.queries)-1 != 3 || len(s.rounds[0]) != 2 || len(s.rounds[1]) != 1 || len(s.rounds[2]) != 1 || s.rounds[2][0].id != 102 {
		t.Fatalf("rounds %v", s.rounds)
	}
	checkPages(t, "A", charIDs(got["1"]), 101, false, 3)
	checkPages(t, "B", charIDs(got["2"]), 102, false, 5)
}

func TestFollowUps_DefaultCapTen(t *testing.T) {
	s := &followUpServer{
		first:  page(bm(101, 1, embedded(101, false, 2, true))),
		answer: func(c fuCall) string { return conn(true, repeatEdges(c, 10)...) },
	}
	got, err := s.client(t).GetByMalIDs(context.Background(), []string{"1"}, types.MediaOpts{CharacterPages: 2})
	if err != nil {
		t.Fatal(err)
	}
	last := s.rounds[len(s.rounds)-1]
	if len(s.rounds) != 8 || last[0].page != 10 {
		t.Fatalf("rounds %v", s.rounds)
	}
	if n := len(got["1"].Characters); n != 2+80 {
		t.Fatalf("chars %d, want 82", n)
	}
}

func repeatEdges(c fuCall, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = cEdge(edgeID(c.id, false, c.page)*100 + i)
	}
	return out
}

func TestFollowUps_CustomCap(t *testing.T) {
	s := &followUpServer{
		first:  page(bm(101, 1, embedded(101, false, 2, true))),
		answer: func(c fuCall) string { return followEdge(c, true) },
	}
	got, err := s.client(t).GetByMalIDs(context.Background(), []string{"1"}, types.MediaOpts{CharacterPages: 2, MaxCharacterPages: 40})
	if err != nil {
		t.Fatal(err)
	}
	if last := s.rounds[len(s.rounds)-1]; len(s.rounds) != 38 || last[0].page != 40 {
		t.Fatalf("rounds %d", len(s.rounds))
	}
	checkPages(t, "chars", charIDs(got["1"]), 101, false, 40)

	s2 := &followUpServer{
		first:  page(bm(101, 1, embedded(101, false, 1, true))),
		answer: func(c fuCall) string { return followEdge(c, true) },
	}
	got, err = s2.client(t).GetByMalIDs(context.Background(), []string{"1"}, types.MediaOpts{CharacterPages: 2, MaxCharacterPages: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s2.queries[0], "c1:") || strings.Contains(s2.queries[0], "c2:") || len(s2.queries) != 1 {
		t.Fatalf("queries %d %s", len(s2.queries), s2.queries[0])
	}
	checkPages(t, "chars", charIDs(got["1"]), 101, false, 1)
}

func TestFollowUps_NullMediaStops(t *testing.T) {
	s := &followUpServer{
		first: page(bm(101, 1, embedded(101, false, 2, true)), bm(102, 2, embedded(102, false, 2, true))),
		answer: func(c fuCall) string {
			if c.id == 101 {
				return "null"
			}
			return followEdge(c, c.page < 4)
		},
	}
	got, err := s.client(t).GetByMalIDs(context.Background(), []string{"1", "2"}, types.MediaOpts{CharacterPages: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.rounds) != 2 || len(s.rounds[1]) != 1 || s.rounds[1][0].id != 102 {
		t.Fatalf("rounds %v", s.rounds)
	}
	checkPages(t, "A", charIDs(got["1"]), 101, false, 2)
	checkPages(t, "B", charIDs(got["2"]), 102, false, 4)
}

func TestFollowUps_ErrorFailsCall(t *testing.T) {
	s := &followUpServer{
		first:  page(bm(101, 1, embedded(101, false, 2, true))),
		answer: func(c fuCall) string { return followEdge(c, true) },
		status: func(n int) int {
			if n >= 2 { // httpx retries a 500, so every follow-up attempt fails
				return http.StatusInternalServerError
			}
			return 0
		},
	}
	if _, err := s.client(t).GetByMalIDs(context.Background(), []string{"1"}, types.MediaOpts{CharacterPages: 2}); err == nil {
		t.Fatal("want error")
	}
}

func TestGetByMalID_UsesPackedPath(t *testing.T) {
	s := &followUpServer{
		first:  `{"data":{"Media":{"id":7,"title":{"romaji":"x"},` + embedded(7, false, 2, true) + `,` + embedded(7, true, 3, true) + `}}}`,
		answer: func(c fuCall) string { return followEdge(c, false) },
	}
	m, err := s.client(t).GetByMalID(context.Background(), "1", types.MediaOpts{CharacterPages: 2, StaffPages: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.queries) != 2 || len(s.rounds) != 1 || len(s.rounds[0]) != 2 {
		t.Fatalf("queries %d rounds %v", len(s.queries), s.rounds)
	}
	checkPages(t, "chars", charIDs(m), 7, false, 3)
	checkPages(t, "staff", staffIDs(m), 7, true, 4)
}

func TestFollowUps_GraphQLErrorFailsCall(t *testing.T) {
	s := &followUpServer{
		first:  page(bm(101, 1, embedded(101, false, 2, true))),
		errs:   `[{"message":"Max query complexity","status":400}]`,
		answer: func(c fuCall) string { return "null" },
	}
	_, err := s.client(t).GetByMalIDs(context.Background(), []string{"1"}, types.MediaOpts{CharacterPages: 2})
	if err == nil || !strings.Contains(err.Error(), "Max query complexity") {
		t.Fatalf("want GraphQL error, got %v", err)
	}
}

func TestFollowUps_MetricsOp(t *testing.T) {
	s := &followUpServer{
		first:  page(bm(101, 1, embedded(101, false, 2, true))),
		answer: func(c fuCall) string { return followEdge(c, false) },
	}
	var ops []string
	c := s.clientWith(t, func(_, op string, _ time.Duration, _ error) { ops = append(ops, op) })
	if _, err := c.GetByMalIDs(context.Background(), []string{"1"}, types.MediaOpts{CharacterPages: 2}); err != nil {
		t.Fatal(err)
	}
	if len(ops) != 2 || ops[0] != OpGetByMalIDs || ops[1] != OpFollowUps {
		t.Fatalf("ops %v", ops)
	}
}

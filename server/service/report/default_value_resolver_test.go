package report

import (
	"testing"
	"time"
)

func TestResolveDefaultValue_Today(t *testing.T) {
	got := ResolveDefaultValue("today")
	if got != time.Now().Format("2006-01-02") {
		t.Fatalf("got %v", got)
	}
}

func TestResolveDefaultValue_ThisMonth(t *testing.T) {
	rv, ok := ResolveDefaultValue("thisMonth").(RangeValue)
	if !ok {
		t.Fatalf("want RangeValue, got %T", ResolveDefaultValue("thisMonth"))
	}
	now := time.Now()
	wantFirst := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	if rv[0] != wantFirst {
		t.Fatalf("start: got %s want %s", rv[0], wantFirst)
	}
	if rv[1] != time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, -1).Format("2006-01-02") {
		t.Fatalf("end: %s", rv[1])
	}
}

func TestResolveDefaultValue_PlainValuePassthrough(t *testing.T) {
	if got := ResolveDefaultValue("2026-01-15"); got != "2026-01-15" {
		t.Fatalf("got %v", got)
	}
}

func TestResolveToMap_DateRangeExpr(t *testing.T) {
	m := map[string]interface{}{}
	ResolveToMap("d", "dateRange", "thisMonth", m)
	if _, ok := m["d_start"]; !ok {
		t.Fatalf("missing d_start: %v", m)
	}
	if _, ok := m["d_end"]; !ok {
		t.Fatalf("missing d_end: %v", m)
	}
}

func TestResolveToMap_DateRangeString(t *testing.T) {
	m := map[string]interface{}{}
	ResolveToMap("d", "dateRange", "2026-01-01,2026-01-31", m)
	if m["d_start"] != "2026-01-01" || m["d_end"] != "2026-01-31" {
		t.Fatalf("got %v", m)
	}
}

func TestResolveToMap_SingleValue(t *testing.T) {
	m := map[string]interface{}{}
	ResolveToMap("d", "date", "today", m)
	if m["d"] != time.Now().Format("2006-01-02") {
		t.Fatalf("got %v", m)
	}
	m2 := map[string]interface{}{}
	ResolveToMap("n", "number", "42", m2)
	if m2["n"] != "42" {
		t.Fatalf("got %v", m2)
	}
}

func TestResolveToMap_EmptySkipped(t *testing.T) {
	m := map[string]interface{}{}
	ResolveToMap("x", "string", "", m)
	if len(m) != 0 {
		t.Fatalf("got %v", m)
	}
}

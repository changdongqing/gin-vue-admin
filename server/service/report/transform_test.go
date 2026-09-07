package report

import (
	"strings"
	"testing"
	"time"
)

func TestJsTransform_MapAndCompute(t *testing.T) {
	in := &QueryResult{Columns: []string{"name", "qty"}, Rows: []map[string]interface{}{
		{"name": "A", "qty": "2"}, {"name": "B", "qty": "3"},
	}}
	out, err := (&JsTransformExecutor{}).Execute(in, "return data.map(r => ({ ...r, qty: Number(r.qty) * 10, tag: r.name + '!' }))")
	if err != nil {
		t.Fatal(err)
	}
	if out.Rows[0]["qty"].(int64) != int64(20) && out.Rows[0]["qty"] != float64(20) {
		t.Fatalf("qty: %v", out.Rows[0]["qty"])
	}
	if out.Rows[1]["tag"] != "B!" {
		t.Fatalf("tag: %v", out.Rows[1]["tag"])
	}
	// 原列序保持在前
	if out.Columns[0] != "name" || out.Columns[1] != "qty" {
		t.Fatalf("columns: %v", out.Columns)
	}
}

func TestJsTransform_TimeoutInterrupt(t *testing.T) {
	in := &QueryResult{Columns: []string{"a"}, Rows: []map[string]interface{}{{"a": 1}}}
	start := time.Now()
	_, err := (&JsTransformExecutor{}).Execute(in, "while(true){}")
	if err == nil || !strings.Contains(err.Error(), "execution timeout") {
		t.Fatalf("expect interrupt error, got %v", err)
	}
	if time.Since(start) > 6*time.Second {
		t.Fatalf("timeout took too long: %v", time.Since(start))
	}
}

func TestJsTransform_NoHostAccess(t *testing.T) {
	in := &QueryResult{Columns: []string{"a"}, Rows: []map[string]interface{}{{"a": 1}}}
	for _, script := range []string{
		"return [{v: typeof require}]",
		"return [{v: typeof java}]",
		"return [{v: typeof Runtime}]",
	} {
		out, err := (&JsTransformExecutor{}).Execute(in, script)
		if err != nil {
			t.Fatalf("script %q err: %v", script, err)
		}
		if out.Rows[0]["v"] != "undefined" {
			t.Fatalf("host object reachable for %q: %v", script, out.Rows[0]["v"])
		}
	}
}

func TestJsTransform_NonArrayReturnRejected(t *testing.T) {
	in := &QueryResult{Columns: []string{"a"}, Rows: []map[string]interface{}{{"a": 1}}}
	if _, err := (&JsTransformExecutor{}).Execute(in, "return 42"); err == nil {
		t.Fatal("expect error for non-array return")
	}
}

func TestDictTransform_Mapping(t *testing.T) {
	in := &QueryResult{Columns: []string{"status", "name"}, Rows: []map[string]interface{}{
		{"status": "0", "name": "x"}, {"status": "2", "name": "y"}, {"status": 1, "name": "z"},
	}}
	out, err := (&DictTransformExecutor{}).Execute(in, `{"field":"status","mapping":{"0":"停用","1":"启用"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if out.Rows[0]["status"] != "停用" {
		t.Fatalf("r0: %v", out.Rows[0])
	}
	if out.Rows[1]["status"] != "2" { // 未命中保持原值
		t.Fatalf("r1: %v", out.Rows[1])
	}
	if out.Rows[2]["status"] != "启用" { // 数值键转字符串命中
		t.Fatalf("r2: %v", out.Rows[2])
	}
}

func TestDictTransform_InvalidJSON(t *testing.T) {
	in := &QueryResult{Columns: []string{"a"}, Rows: []map[string]interface{}{{"a": 1}}}
	if _, err := (&DictTransformExecutor{}).Execute(in, `not-json`); err == nil {
		t.Fatal("expect error")
	}
}

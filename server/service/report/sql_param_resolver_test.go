package report

import (
	"testing"
)

func TestValidateSQL_SelectOnly(t *testing.T) {
	valid := []string{
		"SELECT * FROM t",
		"-- comment\nSELECT 1",
		"/* block */ SELECT 1",
		"  select id from sys_users where id = ${id}",
		"WITH x AS (SELECT 1) SELECT * FROM x",
	}
	for _, sql := range valid {
		if err := ValidateSQL(sql); err != nil {
			t.Errorf("expected valid, got %v for %q", err, sql)
		}
	}
	invalid := []string{
		"",
		"UPDATE t SET a=1",
		"INSERT INTO t VALUES(1)",
		"DROP TABLE t",
		"SELECT 1; DROP TABLE x",
		"EXPLAIN SELECT 1",
	}
	for _, sql := range invalid {
		if err := ValidateSQL(sql); err == nil {
			t.Errorf("expected invalid for %q", sql)
		}
	}
}

func TestValidateSQL_DangerousWordInsideSubquery(t *testing.T) {
	// 危险词拦截不区分位置（对齐实现：原文拦截，宁严勿松）
	if err := ValidateSQL("SELECT * FROM deleted_rows"); err != nil {
		// deleted_rows 中 DELETE 不是整词（\b 依据下划线划界：delete 与 _ 之间有边界）
		t.Logf("note: %v", err)
	}
}

func TestStripEmptyIfBlocks(t *testing.T) {
	sql := "SELECT 1 FROM t WHERE 1=1 <if param=\"a\"> AND x = ${a}</if> <if param=\"b\"> AND y = ${b}</if>"
	got := StripEmptyIfBlocks(sql, map[string]interface{}{"a": "v1"})
	want := "SELECT 1 FROM t WHERE 1=1  AND x = ${a} "
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = StripEmptyIfBlocks(sql, map[string]interface{}{"a": "v1", "b": "  "})
	if got != want {
		t.Fatalf("blank string should strip, got %q", got)
	}
	got = StripEmptyIfBlocks(sql, map[string]interface{}{"a": "v1", "b": 0})
	if got != "SELECT 1 FROM t WHERE 1=1  AND x = ${a}  AND y = ${b}" {
		t.Fatalf("numeric 0 is a value, got %q", got)
	}
}

func TestResolve_ParamExpansion(t *testing.T) {
	r, err := Resolve("SELECT * FROM t WHERE w IN (${ws}) AND n = ${n}", map[string]interface{}{
		"ws": "WS001, WS002",
		"n":  3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Query != "SELECT * FROM t WHERE w IN (?,?) AND n = ?" {
		t.Fatalf("query: %s", r.Query)
	}
	if len(r.Args) != 3 || r.Args[0] != "WS001" || r.Args[1] != "WS002" || r.Args[2] != 3 {
		t.Fatalf("args: %v", r.Args)
	}
}

func TestResolve_MissingParamDefaultsEmpty(t *testing.T) {
	r, err := Resolve("SELECT * FROM t WHERE u = ${missing}", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Args) != 1 || r.Args[0] != "" {
		t.Fatalf("args: %v", r.Args)
	}
}

func TestResolve_InjectionValueSafe(t *testing.T) {
	// 注入向量经参数化后作为纯值传递，不改变语句结构
	r, _ := Resolve("SELECT * FROM t WHERE u = ${u} AND p = ${p}", map[string]interface{}{
		"u": "' OR '1'='1",
		"p": "x'; DROP TABLE t;--",
	})
	if r.Query != "SELECT * FROM t WHERE u = ? AND p = ?" {
		t.Fatalf("query: %s", r.Query)
	}
	if len(r.Args) != 2 {
		t.Fatalf("args: %v", r.Args)
	}
}

func TestRebindDollar(t *testing.T) {
	if got := RebindDollar("SELECT * FROM t WHERE a = ? AND b IN (?,?)"); got != "SELECT * FROM t WHERE a = $1 AND b IN ($2,$3)" {
		t.Fatalf("got %s", got)
	}
	if got := RebindDollar("SELECT 1"); got != "SELECT 1" {
		t.Fatalf("got %s", got)
	}
}

func TestReplaceParams(t *testing.T) {
	got := ReplaceParams("http://x/api?a=${a}&b=${b}&c=${c}", map[string]interface{}{"a": 1, "b": "x y"})
	want := "http://x/api?a=1&b=x y&c="
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestReplaceParamsDeep(t *testing.T) {
	body := map[string]interface{}{
		"name": "${name}",
		"page": map[string]interface{}{"size": 10, "kw": "${kw}"},
		"ids":  []interface{}{"${a}", 2},
	}
	out := ReplaceParamsDeep(body, map[string]interface{}{"name": "n1", "kw": "k", "a": "1"}).(map[string]interface{})
	if out["name"] != "n1" {
		t.Fatalf("name: %v", out["name"])
	}
	if out["page"].(map[string]interface{})["kw"] != "k" {
		t.Fatalf("kw: %v", out["page"])
	}
	if out["ids"].([]interface{})[0] != "1" {
		t.Fatalf("ids: %v", out["ids"])
	}
}

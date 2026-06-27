package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
	"time"
)

// scanFile の各ケースを検証する（now を固定して期限判定を決定的にする）。
func Test_scanFile(t *testing.T) {
	today := time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC)

	src := `package x

// TODO(debt: 2026-01-01): これは期限切れ
func A() {}

// TODO(debt: 2099-01-01): これは未来なのでOK
func B() {}

// TODO(debt): 期限が無い
func C() {}

// TODO(debt: 2026-01-01): これは抑制 //debtlint:allow 来週やる
func D() {}

// ただの TODO: これは負債ではない
func E() {}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "sample.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	got := scanFile(fset, f, today)

	if len(got) != 2 {
		t.Fatalf("違反は 2 件のはず（期限切れ + 期限未設定）、got %d 件: %+v", len(got), got)
	}
	joined := got[0].msg + " | " + got[1].msg
	if !strings.Contains(joined, "期限切れ") {
		t.Errorf("期限切れの違反が含まれていません: %v", got)
	}
	if !strings.Contains(joined, "期限がありません") {
		t.Errorf("期限未設定の違反が含まれていません: %v", got)
	}
}

// ignore-file でファイル全体が無視されること。
func Test_scanFile_ignoreFile(t *testing.T) {
	today := time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC)
	src := `//debtlint:ignore-file 生成ファイルのため
package x

// TODO(debt: 2020-01-01): 期限切れだが ignore-file で無視される
func A() {}
`
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "gen.go", src, parser.ParseComments)
	if got := scanFile(fset, f, today); len(got) != 0 {
		t.Errorf("ignore-file なので 0 件のはず、got %v", got)
	}
}

// runCLI の OK パス（負債コメントが無いツリー）。
func Test_runCLI_ok(t *testing.T) {
	root := t.TempDir()
	var out, errb strings.Builder
	code := runCLI([]string{root}, &out, &errb, time.Now())
	if code != 0 {
		t.Fatalf("空ツリーは exit 0 のはず、got %d (%s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "OK") {
		t.Errorf("OK メッセージが出ていません: %s", out.String())
	}
}

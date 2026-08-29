package database

import (
	_ "embed"
	"fmt"
	"strings"
)

// schemaDDL は実スキーマの正本（`schema/schema.sql`）。
// バイナリに埋め込んで起動時に流すため、デプロイ物とスキーマ定義が必ず同じ版になる。
// 同じファイルが sqlc の型付け入力でもあるので、宣言と実体がずれない。
//
//go:embed schema/schema.sql
var schemaDDL string

// スキーマの節を分ける印。ファイル側の見出しと一字一句そろえること。
//
// 1 ファイルにまとめてあるが、**適用は 2 回に分かれる**。中核（Ⅰ）とノート（Ⅱ+Ⅲ）の
// あいだに seed とバックフィルが挟まり、それらは Ⅰ の表を必要とし、かつ Ⅱ+Ⅲ より
// 先に済んでいる必要があるため（順序の理由は Migrate を読むこと）。
// そこで読み込み時に印で切り、必要な範囲だけを流す。
const (
	coreSectionMarker = "-- Ⅰ. 中核"
	noteSectionMarker = "-- Ⅱ. ノートの骨格"
)

// coreSchemaSection は中核（Ⅰ）だけを返す。
func coreSchemaSection() (string, error) {
	return schemaSection(coreSectionMarker, noteSectionMarker)
}

// noteSchemaSection はノートの骨格と権限（Ⅱ + Ⅲ）を返す。
// Ⅱ と Ⅲ は続けて流してよい（Ⅲ が参照する Ⅰ の users は先に作られている）。
func noteSchemaSection() (string, error) {
	return schemaSection(noteSectionMarker, "")
}

// schemaSection は from の印から to の印の手前までを切り出す。
// to が空なら末尾まで。**印が見つからなければエラーにする** — 見出しを変えたときに
// 黙って空の DDL を流し、「起動は成功したが表が無い」で気づけなくなるのを防ぐ。
func schemaSection(from, to string) (string, error) {
	start := strings.Index(schemaDDL, from)
	if start < 0 {
		return "", fmt.Errorf("スキーマの節が見つかりません（印: %q）。schema/schema.sql の見出しを変えたなら schema_source.go の印も合わせること", from)
	}
	rest := schemaDDL[start:]
	if to == "" {
		return rest, nil
	}
	end := strings.Index(rest, to)
	if end < 0 {
		return "", fmt.Errorf("スキーマの節が見つかりません（印: %q）", to)
	}
	return rest[:end], nil
}

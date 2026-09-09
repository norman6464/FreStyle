// Package pgtext は sqlc.yaml の go_type override 用の小さな database/sql
// Scanner / Valuer 型を置く（ドライバの型変換をそのまま信用できない列専用）。
package pgtext

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// NullDate は PostgreSQL の date 列（NULL 可）を 'YYYY-MM-DD' の文字列として運ぶ。
//
// pgx の stdlib（database/sql）互換層は date 型（OID 1082）を、宛先の Go 型に関わらず
// 内部で pgtype.Date → time.Time にデコードしてから driver.Value として渡す
// （stdlib/sql.go の Rows.Next が date OID をハードコードで pgtype.Date に読み、
// .Value() で time.Time に変換する。TypeMap.RegisterType でこの OID の Codec を
// 差し替えても、この switch 文自体が変数の型を pgtype.Date に固定しているので効かない —
// pgx v5.9.2 で実機確認済み）。simple / extended のどちらの query protocol でも同じ。
//
// 宛先を sql.NullString にしても、その Scan は database/sql の convertAssign を経由し、
// time.Time → *string の変換に time.RFC3339Nano を使うため
// （"2026-09-01" ではなく "2026-09-01T00:00:00Z" になる）、sqlc.yaml の
// date → sql.NullString override だけでは直らない。この型を宛先にすることで、
// time.Time で来ても string で来ても 'YYYY-MM-DD' に揃える。
//
// tickets.start_date / due_date の 2 列専用（設計 Ⅳ-K: date は本番の simple protocol で
// time.Time として運ぶと 1 日ずれるため、そもそも文字列で扱う方針）。
type NullDate struct {
	String string
	Valid  bool
}

// Scan implements sql.Scanner.
func (d *NullDate) Scan(src any) error {
	if src == nil {
		d.String, d.Valid = "", false
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		d.String = v.Format("2006-01-02")
	case string:
		d.String = v
	case []byte:
		d.String = string(v)
	default:
		return fmt.Errorf("pgtext: NullDate.Scan: unsupported source type %T", src)
	}
	d.Valid = true
	return nil
}

// Value implements driver.Valuer.
func (d NullDate) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.String, nil
}

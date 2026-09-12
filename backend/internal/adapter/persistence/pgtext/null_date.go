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
// pgx の stdlib 互換層は date 型（OID 1082）を、宛先の Go 型に関わらず内部で
// pgtype.Date → time.Time にデコードしてから driver.Value として渡す（Rows.Next が OID を
// ハードコードで pgtype.Date に読むため、TypeMap.RegisterType で Codec を差し替えても効かない —
// pgx v5.9.2 で実機確認済み。simple / extended どちらの protocol でも同じ）。
//
// 宛先を sql.NullString にしても、その Scan は convertAssign 経由で time.Time → *string に
// time.RFC3339Nano を使うため（"2026-09-01" ではなく "2026-09-01T00:00:00Z" になる）直らない。
// この型を宛先にすることで、time.Time で来ても string で来ても 'YYYY-MM-DD' に揃える。
//
// tickets.start_date / due_date の 2 列専用（date を time.Time で運ぶと本番の simple protocol
// で 1 日ずれるため、そもそも文字列で扱う方針）。
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

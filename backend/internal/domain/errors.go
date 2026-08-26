package domain

import "errors"

// ErrNotFound は「対象のレコードが存在しない」ことを表す番兵エラー。
// repository が返し、usecase / handler は errors.Is で判定して 404 に分岐する
// （DB ドライバ固有のエラーを上位層へ漏らさないための、層をまたぐ唯一の合図）。
var ErrNotFound = errors.New("domain: record not found")

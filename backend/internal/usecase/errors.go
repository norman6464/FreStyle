package usecase

import "errors"

// 層をまたいで意味が変わらない共通のエラー。
//
// handler はこれらを errors.Is で判定して HTTP ステータスにマップする。
// 文字列比較（err.Error() == "forbidden" 等）はエラー文言を変えた瞬間に
// 静かに 500 へ化けるため使わないこと。
//
// 個別の事情がある場合は fmt.Errorf("%w: 詳細", ErrForbidden) のようにラップして、
// errors.Is での判定を保ったまま文脈を足す。
var (
	// ErrForbidden は認証は済んでいるが操作の権限がないことを表す（403）。
	ErrForbidden = errors.New("forbidden")
)

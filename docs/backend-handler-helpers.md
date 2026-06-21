# backend handler の共通ヘルパー

handler 層で繰り返し書かれていた「認証 actor の取り出し」と「usecase エラーの HTTP 振り分け」を、
[internal/handler/helpers.go](../backend/internal/handler/helpers.go) の共通関数に集約した。

## actorFromContext

```go
func actorFromContext(c *gin.Context) (userID, companyID uint64, role string, ok bool)
```

middleware が注入した current user（[CurrentUserFromContext](../backend/internal/handler/middleware/current_user.go)）から
`userID / companyID / role` を取り出す。未認証なら 401(`unauthorized`) を書いて `ok=false` を返すので、
呼び出し側は `if !ok { return }` で早期 return する。

以前は各 handler が `actorContext` メソッドを個別に実装していた（teaching_material / course で完全重複）。
これを撤去し共通関数へ統一した。

## respondEntityErr

```go
func respondEntityErr(c *gin.Context, err error, notFoundMsg, fallback string)
```

usecase が返したエラーを HTTP ステータスへ振り分ける:

- `gorm.ErrRecordNotFound` → 404（`notFoundMsg`）
- `forbidden*` で始まる / `actor must belong to a company` → 403（`操作権限がありません`）
- それ以外 → 500（`fallback`）

エンティティ固有の文言は `notFoundMsg` / `fallback` で渡す（例: 「教材が見つかりません」/「教材の取得に失敗しました」）。
以前は `respondTeachingMaterialErr` / `respondCourseErr` がほぼ同一実装で重複していた。

> 注: 認可エラーを文字列（`forbidden...`）で判定しているのは既存挙動の踏襲。将来的には usecase 側で
> sentinel error（`var ErrForbidden = errors.New(...)`）を定義し `errors.Is` で判定するのが望ましい（別 PR）。

## domain.User.CompanyIDValue()

```go
func (u User) CompanyIDValue() uint64 // CompanyID が nil なら 0
```

`companyID := uint64(0); if user.CompanyID != nil { companyID = *user.CompanyID }` の展開を 1 行にする
domain の小道具（純粋関数・副作用なし）。actorFromContext や chapter_view / lesson_progress handler で利用。

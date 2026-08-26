// Package domain は他層に依存しない純粋なドメイン構造体・列挙・定数を集める。
// DB スキーマは domain から起こさず、infra/database/schema/*.sql の明示 DDL を正本とする
// （同じファイルが sqlc の型付け入力でもある）。
package domain

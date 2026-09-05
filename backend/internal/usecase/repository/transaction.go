package repository

import "context"

// TxManager は複数 repository の書き込みを1つのトランザクションにまとめる。
//
// fn に渡す ctx にトランザクションが埋め込まれるため、fn の中で呼ぶ repository は
// 何も意識せずにこのトランザクションへ乗る（repository のメソッドシグネチャは変わらない）。
// fn がエラーを返す（または panic する）と、fn の中で行った書き込みはすべてロールバックされる。
type TxManager interface {
	DoInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

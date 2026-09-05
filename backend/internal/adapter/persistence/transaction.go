package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// 非公開の構造体型をキーにすることで、他パッケージの context.WithValue と衝突しない。
type txKeyType struct{}

var txKey = txKeyType{}

// withTx は tx を ctx に埋め込む。sqlTxManager.DoInTx が呼び出す。
func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// getTx は ctx から tx を取り出す。各 repository の baseRepository.dbtx が呼び出す。
func getTx(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey).(*sql.Tx)
	return tx, ok
}

type sqlTxManager struct{ db *sql.DB }

// NewTxManager は repository.TxManager の実装を組み立てる。
func NewTxManager(db *sql.DB) repository.TxManager {
	return &sqlTxManager{db: db}
}

// DoInTx はトランザクションを開始して fn を実行する。fn に渡す ctx にトランザクションが
// 埋め込まれるため、fn の中で呼ぶ repository は baseRepository.dbtx 経由で自動的にこの
// トランザクションへ乗る。fn がエラーを返せば Rollback、成功すれば Commit する。
//
// panic した場合も Rollback してから再送出する（tx を開いたまま接続をリークさせない）。
func (tm *sqlTxManager) DoInTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	// 既にこの ctx がトランザクションの中なら、新規に開始せず今の tx へ相乗りする
	// （あるusecaseが別のusecase/ヘルパーを呼び、そちらも DoInTx しようとするケースへの備え。
	// 二重に BeginTx するとデッドロックの原因になる。commit/rollback は外側の DoInTx だけが持つ）。
	if _, inTx := getTx(ctx); inTx {
		return fn(ctx)
	}

	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(withTx(ctx, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback: %w, original: %w", rbErr, err)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

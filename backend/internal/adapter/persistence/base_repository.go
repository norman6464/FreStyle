package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
)

// baseRepository は db への接続保持と、トランザクション対応のクエリ実行先解決を
// 全 repository で共有する。各 repository はこれを埋め込むだけで dbtx(ctx) が使えるようになる。
type baseRepository struct {
	db *sql.DB
}

// dbtx は ctx に進行中のトランザクション（sqlTxManager.DoInTx が埋め込んだもの）があれば
// それを、無ければ通常の *sql.DB を返す。repository の全メソッドがこれ経由でクエリ実行先を
// 得ることで、トランザクションの有無によらず同じ実装を使い回せる。
//
// sqlcgen.DBTX は *sql.DB と *sql.Tx のどちらも満たすので、呼び出し側は
// sqlcgen.New(r.dbtx(ctx)) とするだけでよい。
func (b *baseRepository) dbtx(ctx context.Context) sqlcgen.DBTX {
	if tx, ok := getTx(ctx); ok {
		return tx
	}
	return b.db
}

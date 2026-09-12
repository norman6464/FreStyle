// rebuildsearchindex は既存の全ページの page_search / page_links（本文検索・逆リンク）を
// 作り直す一回限りの再構築コマンド。全ワークスペースの全アーカイブ済みでないページについて
// knowledgeBaseRepository.RebuildPageSearchAndLinks を呼び、その時点の blocks から DELETE +
// UPSERT で書き直す（冪等——何度実行しても結果は同じ）。
//
// 通常の保存経路（ReplacePageBlocksUseCase）は保存のたびに自動で page_search / page_links を
// 同期するので、このコマンドが要るのは「この機能が入る前から存在していたページ」を一度だけ
// 追いつかせるためだけ。流し忘れても本文保存やページ表示自体は壊れない（page_search /
// page_links は blocks から再生成できる派生データなので、検索や逆リンクに一時的に
// 出てこないだけで済む）。
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/infra/config"
	"github.com/norman6464/frestyle/backend/internal/infra/database"
	"github.com/norman6464/frestyle/backend/internal/infra/logging"
)

// fatal は cmd/server/main.go の fatal と同じ形（致命的エラーを構造化ログで出して終了する）。
func fatal(msg string, err error) {
	slog.Error(msg, slog.Any("error", err))
	os.Exit(1)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		logging.Setup("")
		fatal("config load failed", err)
	}
	logging.Setup(cfg.AppEnv)

	// cmd/server/main.go と同じ接続の作り方（simple query protocol の強制分岐も含め
	// NewPostgres 内で共通化されている）。
	sqlDB, err := database.NewPostgres(cfg)
	if err != nil {
		fatal("database connect failed", err)
	}
	defer func() {
		if cerr := sqlDB.Close(); cerr != nil {
			slog.Error("database close failed", slog.Any("error", cerr))
		}
	}()

	repo := persistence.NewKnowledgeBaseRepository(sqlDB)
	ctx := context.Background()

	workspaceIDs, err := repo.ListAllWorkspaceIDs(ctx)
	if err != nil {
		fatal("list workspaces failed", err)
	}
	slog.Info("rebuildsearchindex: starting", slog.Int("workspaces", len(workspaceIDs)))

	var (
		totalPages           int
		totalOK              int
		totalErr             int
		failedWorkspaceCount int
	)
	for _, workspaceID := range workspaceIDs {
		pageIDs, err := repo.ListActivePageIDsByWorkspace(ctx, workspaceID)
		if err != nil {
			// 1 ワークスペースの列挙失敗で全体を止めない。件数を記録して次へ進む。
			failedWorkspaceCount++
			slog.Error("rebuildsearchindex: list pages failed",
				slog.String("workspaceId", workspaceID), slog.Any("error", err))
			continue
		}
		slog.Info("rebuildsearchindex: workspace",
			slog.String("workspaceId", workspaceID), slog.Int("pages", len(pageIDs)))

		for _, pageID := range pageIDs {
			totalPages++
			if err := repo.RebuildPageSearchAndLinks(ctx, workspaceID, pageID); err != nil {
				totalErr++
				// 1 ページの失敗で全体を止めない（同じ理由 —
				// 後続の正常なページまで巻き込んで再構築を止めたくない）。
				slog.Error("rebuildsearchindex: page failed",
					slog.String("workspaceId", workspaceID), slog.String("pageId", pageID), slog.Any("error", err))
				continue
			}
			totalOK++
			// 進捗（件数）を標準出力に出す。大量のページを流すことを想定し、
			// 1 件ごとではなく一定件数ごとに出す（ログが埋もれないように）。
			if totalPages%100 == 0 {
				slog.Info("rebuildsearchindex: progress",
					slog.Int("processed", totalPages), slog.Int("ok", totalOK), slog.Int("failed", totalErr))
			}
		}
	}

	slog.Info("rebuildsearchindex: done",
		slog.Int("workspaces", len(workspaceIDs)),
		slog.Int("failedWorkspaces", failedWorkspaceCount),
		slog.Int("pages", totalPages), slog.Int("ok", totalOK), slog.Int("failed", totalErr))
	if totalErr > 0 || failedWorkspaceCount > 0 {
		// 1 件でも失敗したページ・ワークスペース列挙があれば、CI・運用のバッチ監視が
		// 気づけるよう非 0 で終了する。
		os.Exit(1)
	}
}

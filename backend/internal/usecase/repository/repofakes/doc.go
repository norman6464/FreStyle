// Package repofakes は repository interface(port) のテスト用 fake を提供する。
//
// fakes_gen.go は cmd/fakegen が自動生成する（手で編集しない）。port にメソッドが増えたら
// 再生成するだけで fake が追従する。usecase の単体テストで、本物の DB/外部サービスの代わりに
// この fake を注入する。
//
//	repo := &repofakes.FakeLessonProgressRepository{
//	    MarkCompletedFunc: func(ctx context.Context, userID, materialID, courseID uint64) (bool, error) {
//	        return true, nil
//	    },
//	}
//	uc := usecase.NewMarkLessonCompletedUseCase(repo, ...)
//	// ... repo.MarkCompletedCalls で呼び出し回数も検証できる
//
// 再生成: make fakegen （= go run ./cmd/fakegen . && gofumpt -w ./internal/usecase/repository/repofakes）
//
//go:generate go run ../../../../cmd/fakegen ../../../..
package repofakes

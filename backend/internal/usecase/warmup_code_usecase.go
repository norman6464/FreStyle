package usecase

import "context"

// WarmupCodeUseCase は指定言語の実行環境を事前に温める。学習者がコードエディタ
// （演習詳細）ページに入った時点で呼び、最初の Run を即時化する用途。
type WarmupCodeUseCase struct {
	runner CodeRunner
}

// NewWarmupCodeUseCase は CodeRunner を注入して WarmupCodeUseCase を返す。
func NewWarmupCodeUseCase(runner CodeRunner) *WarmupCodeUseCase {
	return &WarmupCodeUseCase{runner: runner}
}

// Execute は language の実行環境を温める（Go はコンパイルキャッシュ、php/bash は no-op）。
func (uc *WarmupCodeUseCase) Execute(ctx context.Context, language string) error {
	return uc.runner.Warmup(ctx, language)
}

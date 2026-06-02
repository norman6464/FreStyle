package usecase

import "github.com/norman6464/FreStyle/backend/internal/domain"

// SendAiMessageStreamUseCase 用の共通型・ヘルパを置く。

// SendAiMessageInput は AI チャット送信リクエスト。
// Scene / SessionType / ScenarioID は旧仕様の互換で残すが、プロンプトには反映しない。
// Attachments は S3 へ presigned PUT 済みの key + メタで、usecase が実体を取得して Bedrock に渡す。
type SendAiMessageInput struct {
	UserID      uint64
	SessionID   uint64
	Content     string
	Scene       string
	SessionType string
	ScenarioID  *uint64
	Attachments []domain.Attachment
}

func truncateTitle(s string, max int) string {
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max]) + "…"
	}
	return s
}

// buildSystemPrompt は汎用 AI チャットの system prompt を返す（引数は旧互換で未使用）。
func buildSystemPrompt(_, _ string) string {
	return "あなたは新卒 IT エンジニア向け学習プラットフォーム『FreStyle』に組み込まれた汎用 AI アシスタントです。" +
		"ユーザーからの質問・要約依頼・コードレビュー・概念説明など、幅広いトピックに答えます。" +
		"日本語で簡潔・丁寧に応答してください。" +
		"回答は **Markdown 形式** で返してください（見出し / 箇条書き / 表 / コードブロックなど、内容に応じて適切に使い分けます）。" +
		"コードを示すときは ```言語名 で囲み、シンタックスハイライトが効くようにしてください。"
}

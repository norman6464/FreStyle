package usecase

import "github.com/norman6464/FreStyle/backend/internal/domain"

// 旧 SendAiMessageUseCase（WebSocket 経由の同期呼び出し）は PR-D で撤去した。
// SSE ストリーミング版（SendAiMessageStreamUseCase）が後継。
//
// 共通で使う型・ヘルパだけここに残す:
//   - SendAiMessageInput: stream usecase の Execute 引数
//   - truncateTitle:      新規セッション作成時のタイトル切り詰め
//   - buildSystemPrompt:  Bedrock に渡す汎用 AI アシスタント用 system prompt

// SendAiMessageInput は AI チャット送信リクエスト。
//
// Scene / SessionType / ScenarioID は旧フロント仕様の互換のため残しているが、
// 練習モード撤去（PR-A）以降はどれもプロンプトに反映しない。次のクリーンアップで削除予定。
//
// Attachments はユーザー発話に紐付く添付ファイル群（画像 / 将来の PDF / CSV）。
// フロントは事前に S3 へ presigned PUT でアップロード済みで、ここには key と
// メタデータだけが渡る。usecase が S3 から実体バイトを取得して Bedrock に渡す。
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

// buildSystemPrompt は汎用 AI チャットの system prompt を返す。
//
// 引数 sessionType / scene は旧シグネチャ互換で受けるが本文には反映しない。
func buildSystemPrompt(_, _ string) string {
	return "あなたは新卒 IT エンジニア向け学習プラットフォーム『FreStyle』に組み込まれた汎用 AI アシスタントです。" +
		"ユーザーからの質問・要約依頼・コードレビュー・概念説明など、幅広いトピックに答えます。" +
		"日本語で簡潔・丁寧に応答してください。" +
		"回答は **Markdown 形式** で返してください（見出し / 箇条書き / 表 / コードブロックなど、内容に応じて適切に使い分けます）。" +
		"コードを示すときは ```言語名 で囲み、シンタックスハイライトが効くようにしてください。"
}

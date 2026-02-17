package com.example.FreStyle.service;

import org.springframework.stereotype.Component;

/**
 * ビジネスコミュニケーション評価のシステムプロンプトを構築するクラス
 */
@Component
public class SystemPromptBuilder {

    /**
     * 言い換え提案用のシステムプロンプトを構築する
     * 3パターン（フォーマル版/ソフト版/簡潔版）の言い換えを提案する
     *
     * @param scene シーン識別子（null可）
     */
    public String buildRephrasePrompt(String scene) {
        StringBuilder sb = new StringBuilder();

        sb.append("あなたはビジネスコミュニケーションの言い換え提案の専門家です。\n");
        sb.append("ユーザーが入力したメッセージを、以下の5パターンで言い換えてください。\n\n");

        sb.append("【言い換え5パターン】\n");
        sb.append("1. フォーマル版: ビジネスライクで丁寧な表現。敬語を正しく使い、格式のある文体にする\n");
        sb.append("2. ソフト版: 柔らかく配慮のある表現。クッション言葉を活用し、相手の気持ちに寄り添う文体にする\n");
        sb.append("3. 簡潔版: 結論ファーストで短く伝わる表現。冗長な部分を削ぎ落とし、要点のみ伝える文体にする\n");
        sb.append("4. 質問型: 相手に考えを促す問いかけ形式。指示や断定を避け、相手の意見を引き出す表現にする\n");
        sb.append("5. 提案型: 「〜はいかがでしょうか」形式。選択肢を提示し、相手に判断を委ねる丁寧な提案表現にする\n\n");

        if (scene != null && !scene.isEmpty()) {
            sb.append("【シーンの文脈】\n");
            switch (scene) {
                case "meeting":
                    sb.append("会議での発言として適切な言い換えを提案してください。\n\n");
                    break;
                case "one_on_one":
                    sb.append("1on1面談での発言として適切な言い換えを提案してください。\n\n");
                    break;
                case "email":
                    sb.append("ビジネスメールとして適切な言い換えを提案してください。\n\n");
                    break;
                case "presentation":
                    sb.append("プレゼンテーションでの発言として適切な言い換えを提案してください。\n\n");
                    break;
                case "negotiation":
                    sb.append("商談での発言として適切な言い換えを提案してください。\n\n");
                    break;
                case "code_review":
                    sb.append("コードレビューコメントとして適切な言い換えを提案してください。\n\n");
                    break;
                case "incident":
                    sb.append("障害対応の報告・連絡として適切な言い換えを提案してください。\n\n");
                    break;
                case "daily_report":
                    sb.append("日報・週報の記述として適切な言い換えを提案してください。\n\n");
                    break;
                default:
                    break;
            }
        }

        sb.append("【出力形式】\n");
        sb.append("必ず以下のJSON形式で出力してください。JSON以外の文字は一切含めないでください。\n");
        sb.append("{\n");
        sb.append("  \"formal\": \"フォーマル版の言い換え文\",\n");
        sb.append("  \"soft\": \"ソフト版の言い換え文\",\n");
        sb.append("  \"concise\": \"簡潔版の言い換え文\",\n");
        sb.append("  \"questioning\": \"質問型の言い換え文\",\n");
        sb.append("  \"proposal\": \"提案型の言い換え文\"\n");
        sb.append("}\n");

        return sb.toString();
    }

    /**
     * 練習モード用のシステムプロンプトを構築する
     * AIがビジネスシーンの相手役を演じ、ユーザーが対応を練習する
     *
     * @param scenarioName シナリオ名
     * @param roleName 相手役の名前・説明
     * @param difficulty 難易度（beginner, intermediate, advanced）
     * @param scenarioContext シナリオの状況説明
     */
    public String buildPracticePrompt(String scenarioName, String roleName, String difficulty, String scenarioContext) {
        StringBuilder sb = new StringBuilder();

        sb.append("あなたはビジネスコミュニケーションのロールプレイ練習の相手役です。\n");
        sb.append("以下のシナリオ設定に基づいて、相手役を演じてください。\n\n");

        sb.append("【シナリオ】").append(scenarioName).append("\n");
        sb.append("【あなたの役割】").append(roleName).append("\n");

        sb.append("【難易度】");
        switch (difficulty) {
            case "beginner":
                sb.append("初級 - 比較的穏やかに対応し、ユーザーが練習しやすい雰囲気を作ってください\n");
                break;
            case "advanced":
                sb.append("上級 - 厳しい質問や予想外の反応を織り交ぜ、高度な対応力を試してください\n");
                break;
            default:
                sb.append("中級 - 現実的な反応をしつつ、適度にチャレンジングな対応をしてください\n");
                break;
        }

        sb.append("【状況】").append(scenarioContext).append("\n\n");

        sb.append("【指示】\n");
        sb.append("- 設定された役割になりきって自然に会話してください\n");
        sb.append("- ユーザーの発言に対してリアルな反応を返してください\n");
        sb.append("- 一度に長く話しすぎず、会話のキャッチボールを意識してください\n");
        sb.append("- ユーザーが「練習終了」と言ったら、ロールプレイを終了してフィードバックを行ってください\n\n");

        sb.append("【練習終了時のスコア評価】\n");
        sb.append("ユーザーが「練習終了」と入力した場合、このシナリオに応じた評価軸（3〜5軸程度）でスコアリングしてください。\n");
        sb.append("フィードバックの最後に、以下のJSON形式でスコアを出力してください。必ず```jsonと```で囲んでください。\n");
        sb.append("```json\n");
        sb.append("{\"scores\":[\n");
        sb.append("  {\"axis\":\"評価軸1の名称\",\"score\":8,\"comment\":\"一言コメント\"},\n");
        sb.append("  {\"axis\":\"評価軸2の名称\",\"score\":7,\"comment\":\"一言コメント\"},\n");
        sb.append("  {\"axis\":\"評価軸3の名称\",\"score\":6,\"comment\":\"一言コメント\"}\n");
        sb.append("]}\n");
        sb.append("```\n");

        return sb.toString();
    }

    /**
     * 通常チャットモード用のシステムプロンプトを構築する
     * ビジネスコミュニケーションコーチとして応答する
     */
    public String buildCoachPrompt() {
        StringBuilder sb = new StringBuilder();

        sb.append("あなたはビジネスコミュニケーションのプロフェッショナルコーチです。\n");
        sb.append("職場での「伝わる話し方」の専門家として、ユーザーの質問や相談に答えてください。\n\n");

        sb.append("【対象ユーザー】\n");
        sb.append("新卒ITエンジニアを主な対象としています。顧客折衝やシニアエンジニアとのやり取りで求められるビジネスコミュニケーションスキルの習得を支援してください。\n\n");

        sb.append("あなたの強み:\n");
        sb.append("- 論理的で分かりやすい構成で情報を整理し伝える力\n");
        sb.append("- 相手への配慮を忘れない丁寧な表現と、ポジティブな言い換えスキル\n");
        sb.append("- 問題に対して具体的な解決策や代替案を提案する力\n");
        sb.append("- 質問と傾聴を通じて相手の真のニーズを引き出す力\n\n");

        sb.append("回答する際は、上記のスキルを自然に活用しながら、ユーザーにとって分かりやすく丁寧に答えてください。\n");
        sb.append("ビジネスコミュニケーションに関する質問には、実践的な経験に基づいたアドバイスを提供してください。");

        return sb.toString();
    }
}

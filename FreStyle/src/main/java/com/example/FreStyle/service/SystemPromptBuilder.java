package com.example.FreStyle.service;

import org.springframework.stereotype.Component;

/**
 * ビジネスコミュニケーション評価のシステムプロンプトを構築するクラス
 */
@Component
public class SystemPromptBuilder {

    /**
     * フィードバックモード用のシステムプロンプトを構築する
     * ビジネスコミュニケーション5軸で会話を評価する
     */
    public String buildFeedbackPrompt(
            String displayName,
            String selfIntroduction,
            String communicationStyle,
            String personalityTraits,
            String goals,
            String concerns,
            String preferredFeedbackStyle) {

        StringBuilder sb = new StringBuilder();

        sb.append("あなたはビジネスコミュニケーションのプロフェッショナルであり、職場での対話を分析しフィードバックを行う専門家です。\n");
        sb.append("以下の5つの評価軸に基づいて、ユーザーのチャット内容を分析し、具体的で実践的なフィードバックを行ってください。\n\n");

        sb.append("【ビジネスコミュニケーション 5つの評価軸】\n\n");

        sb.append("1. 論理的構成力\n");
        sb.append("   - PREP法（Point→Reason→Example→Point）に沿った構成になっているか\n");
        sb.append("   - 論点が明確で、根拠と主張の関係が分かりやすいか\n");
        sb.append("   - 情報の優先順位が適切に整理されているか\n");
        sb.append("   【IT新卒向けサブ基準】報連相の構造化: 結論→経緯→依頼事項の順で伝えられているか\n\n");

        sb.append("2. 配慮表現\n");
        sb.append("   - クッション言葉（「恐れ入りますが」「お手数ですが」等）を適切に使えているか\n");
        sb.append("   - 否定的な表現を肯定的に言い換えているか（ポジティブ変換）\n");
        sb.append("   - 相手の立場や感情に配慮した表現ができているか\n");
        sb.append("   【IT新卒向けサブ基準】敬語の正確さ: 尊敬語/謙譲語/丁寧語を正しく使い分けられているか\n\n");

        sb.append("3. 要約力\n");
        sb.append("   - 結論が最初に来ているか（結論ファースト）\n");
        sb.append("   - 冗長な表現を避け、簡潔に要点を伝えられているか\n");
        sb.append("   - 相手が一読で内容を把握できる構成になっているか\n");
        sb.append("   【IT新卒向けサブ基準】技術説明の平易化: 非エンジニアにも分かる言葉で技術的内容を説明できているか\n\n");

        sb.append("4. 提案力\n");
        sb.append("   - 問題提起だけでなく、具体的な解決策や代替案を提示しているか\n");
        sb.append("   - 「できません」ではなく「〜であれば可能です」のような建設的な表現ができているか\n");
        sb.append("   - 次のアクションが明確に示されているか\n");
        sb.append("   【IT新卒向けサブ基準】エスカレーション判断力: 自力解決すべきか上位者に相談すべきかの判断が適切か\n\n");

        sb.append("5. 質問・傾聴力\n");
        sb.append("   - 相手の発言を確認・復唱して正しく理解しようとしているか\n");
        sb.append("   - 質問で相手の意図やニーズを深掘りしているか\n");
        sb.append("   - 一方的に話さず、相手に発言の機会を与えているか\n");
        sb.append("   【IT新卒向けサブ基準】要件確認の網羅性: 5W1Hで仕様や要件を漏れなく確認できているか\n\n");

        sb.append("【フィードバック形式】\n");
        sb.append("各評価軸について、具体的な会話箇所を引用しながらフィードバックしてください。\n");
        sb.append("良い点は積極的に褒め、改善点には具体的な言い換え例を必ず提示してください。\n\n");

        sb.append("【スコア出力】\n");
        sb.append("フィードバックの最後に、以下のJSON形式で各軸のスコア（1〜10）を出力してください。\n");
        sb.append("必ず```jsonと```で囲んでください。\n");
        sb.append("```json\n");
        sb.append("{\"scores\":[\n");
        sb.append("  {\"axis\":\"論理的構成力\",\"score\":8,\"comment\":\"一言コメント\"},\n");
        sb.append("  {\"axis\":\"配慮表現\",\"score\":7,\"comment\":\"一言コメント\"},\n");
        sb.append("  {\"axis\":\"要約力\",\"score\":6,\"comment\":\"一言コメント\"},\n");
        sb.append("  {\"axis\":\"提案力\",\"score\":5,\"comment\":\"一言コメント\"},\n");
        sb.append("  {\"axis\":\"質問・傾聴力\",\"score\":9,\"comment\":\"一言コメント\"}\n");
        sb.append("]}\n");
        sb.append("```\n\n");

        // UserProfile情報を動的に埋め込む
        sb.append("【ユーザープロフィール】\n");

        if (displayName != null && !displayName.isEmpty()) {
            sb.append("- 名前: ").append(displayName).append("\n");
        }
        if (selfIntroduction != null && !selfIntroduction.isEmpty()) {
            sb.append("- 自己紹介: ").append(selfIntroduction).append("\n");
        }
        if (communicationStyle != null && !communicationStyle.isEmpty()) {
            sb.append("- コミュニケーションスタイル: ").append(communicationStyle).append("\n");
        }
        if (personalityTraits != null && !personalityTraits.isEmpty()) {
            sb.append("- 性格特性: ").append(personalityTraits).append("\n");
        }
        if (goals != null && !goals.isEmpty()) {
            sb.append("- 目標: ").append(goals).append("\n");
        }
        if (concerns != null && !concerns.isEmpty()) {
            sb.append("- 懸念事項: ").append(concerns).append("\n");
        }
        if (preferredFeedbackStyle != null && !preferredFeedbackStyle.isEmpty()) {
            sb.append("- フィードバックスタイル: ").append(preferredFeedbackStyle).append("\n");
        }

        sb.append("\n上記のプロフィールを踏まえ、このユーザーの目標達成に役立つフィードバックを行ってください。");

        return sb.toString();
    }

    /**
     * シーン別フィードバックモード用のシステムプロンプトを構築する
     * 基本5軸＋シーン固有の追加評価観点を含める
     *
     * @param scene シーン識別子（meeting, one_on_one, email, presentation, negotiation）
     */
    public String buildFeedbackPromptWithScene(
            String scene,
            String displayName,
            String selfIntroduction,
            String communicationStyle,
            String personalityTraits,
            String goals,
            String concerns,
            String preferredFeedbackStyle) {

        // 基本5軸のフィードバックプロンプトを構築
        String basePrompt = buildFeedbackPrompt(
                displayName, selfIntroduction, communicationStyle,
                personalityTraits, goals, concerns, preferredFeedbackStyle);

        if (scene == null || scene.isEmpty()) {
            return basePrompt;
        }

        StringBuilder sb = new StringBuilder(basePrompt);
        sb.append("\n\n【シーン別追加評価観点】\n");

        switch (scene) {
            case "meeting":
                sb.append("このチャットは「会議」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- 発言のタイミング: 適切なタイミングで発言できているか、議論の流れを妨げていないか\n");
                sb.append("- 議論の建設性: 批判だけでなく、建設的な意見や代替案を出しているか\n");
                sb.append("- ファシリテーション: 他の参加者の意見を引き出し、議論を前進させているか\n");
                break;
            case "one_on_one":
                sb.append("このチャットは「1on1」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- 心理的安全性: 相手が安心して本音を話せる雰囲気を作れているか\n");
                sb.append("- フィードバックの具体性: 抽象的でなく、具体的な行動や事例に基づいたフィードバックができているか\n");
                sb.append("- 傾聴の深さ: 表面的な内容だけでなく、相手の感情や本質的な課題に寄り添えているか\n");
                break;
            case "email":
                sb.append("このチャットは「メール」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- 件名の明確さ: メールの目的が件名から即座に分かるか\n");
                sb.append("- 構成の読みやすさ: 段落分け、箇条書き、適切な改行で読みやすい構成になっているか\n");
                sb.append("- アクション明示: 相手に期待するアクション（返信、承認、確認など）が明確に示されているか\n");
                break;
            case "presentation":
                sb.append("このチャットは「プレゼン」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- ストーリー構成: 導入→本題→結論の流れが論理的で引き込まれる構成か\n");
                sb.append("- 聞き手への配慮: 専門用語の説明、具体例の使用など、聞き手のレベルに合わせた説明ができているか\n");
                sb.append("- 質疑応答力: 質問に対して的確かつ簡潔に回答できているか\n");
                break;
            case "negotiation":
                sb.append("このチャットは「商談」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- ニーズヒアリング: 相手の課題やニーズを的確に把握するための質問ができているか\n");
                sb.append("- 価値提案: 自社の商品・サービスの価値を相手のニーズに結びつけて説明できているか\n");
                sb.append("- クロージング: 次のステップや意思決定を促す適切な提案ができているか\n");
                break;
            case "code_review":
                sb.append("このチャットは「コードレビュー」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- 指摘の具体性: 問題箇所と理由を具体的に説明できているか\n");
                sb.append("- 代替案の提示: 指摘だけでなく、改善コード例や代替アプローチを提案できているか\n");
                sb.append("- 相手への配慮: レビューイの成長を意識した建設的なトーンで伝えられているか\n");
                break;
            case "incident":
                sb.append("このチャットは「障害対応」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- 状況報告の正確さ: 影響範囲・発生時刻・現在のステータスを正確に伝えられているか\n");
                sb.append("- エスカレーション判断: 適切なタイミングで上位者やチームに報告・相談できているか\n");
                sb.append("- 事後報告の構成: 原因・対応・再発防止策を整理して報告できているか\n");
                break;
            case "daily_report":
                sb.append("このチャットは「日報・週報」でのコミュニケーションです。基本5軸に加え、以下の観点でも評価してください。\n\n");
                sb.append("- 成果の定量化: 作業内容や進捗を数値や具体的な成果物で示せているか\n");
                sb.append("- 課題の明確化: 直面している課題やブロッカーを明確に記述できているか\n");
                sb.append("- ネクストアクション: 翌日・翌週の計画や優先事項が具体的に示されているか\n");
                break;
            default:
                // 未知のシーンの場合はシーン別観点を追加しない
                break;
        }

        return sb.toString();
    }

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

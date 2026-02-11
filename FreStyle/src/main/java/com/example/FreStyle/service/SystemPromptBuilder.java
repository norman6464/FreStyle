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
        sb.append("   - 情報の優先順位が適切に整理されているか\n\n");

        sb.append("2. 配慮表現\n");
        sb.append("   - クッション言葉（「恐れ入りますが」「お手数ですが」等）を適切に使えているか\n");
        sb.append("   - 否定的な表現を肯定的に言い換えているか（ポジティブ変換）\n");
        sb.append("   - 相手の立場や感情に配慮した表現ができているか\n\n");

        sb.append("3. 要約力\n");
        sb.append("   - 結論が最初に来ているか（結論ファースト）\n");
        sb.append("   - 冗長な表現を避け、簡潔に要点を伝えられているか\n");
        sb.append("   - 相手が一読で内容を把握できる構成になっているか\n\n");

        sb.append("4. 提案力\n");
        sb.append("   - 問題提起だけでなく、具体的な解決策や代替案を提示しているか\n");
        sb.append("   - 「できません」ではなく「〜であれば可能です」のような建設的な表現ができているか\n");
        sb.append("   - 次のアクションが明確に示されているか\n\n");

        sb.append("5. 質問・傾聴力\n");
        sb.append("   - 相手の発言を確認・復唱して正しく理解しようとしているか\n");
        sb.append("   - 質問で相手の意図やニーズを深掘りしているか\n");
        sb.append("   - 一方的に話さず、相手に発言の機会を与えているか\n\n");

        sb.append("【フィードバック形式】\n");
        sb.append("各評価軸について、具体的な会話箇所を引用しながらフィードバックしてください。\n");
        sb.append("良い点は積極的に褒め、改善点には具体的な言い換え例を必ず提示してください。\n\n");

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
        sb.append("ユーザーが入力したメッセージを、以下の3パターンで言い換えてください。\n\n");

        sb.append("【言い換え3パターン】\n");
        sb.append("1. フォーマル版: ビジネスライクで丁寧な表現。敬語を正しく使い、格式のある文体にする\n");
        sb.append("2. ソフト版: 柔らかく配慮のある表現。クッション言葉を活用し、相手の気持ちに寄り添う文体にする\n");
        sb.append("3. 簡潔版: 結論ファーストで短く伝わる表現。冗長な部分を削ぎ落とし、要点のみ伝える文体にする\n\n");

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
                default:
                    break;
            }
        }

        sb.append("【出力形式】\n");
        sb.append("必ず以下のJSON形式で出力してください。JSON以外の文字は一切含めないでください。\n");
        sb.append("{\n");
        sb.append("  \"formal\": \"フォーマル版の言い換え文\",\n");
        sb.append("  \"soft\": \"ソフト版の言い換え文\",\n");
        sb.append("  \"concise\": \"簡潔版の言い換え文\"\n");
        sb.append("}\n");

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

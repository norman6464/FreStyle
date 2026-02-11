package com.example.FreStyle.service;

import org.springframework.stereotype.Component;

/**
 * コールセンター式QA観点のシステムプロンプトを構築するクラス
 */
@Component
public class SystemPromptBuilder {

    /**
     * フィードバックモード用のシステムプロンプトを構築する
     * コールセンターのQA（品質管理）5軸で会話を評価する
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

        sb.append("あなたはコールセンターの品質管理（QA）のプロフェッショナルであり、コミュニケーションのフィードバックを行う専門家です。\n");
        sb.append("以下の5つの評価軸に基づいて、ユーザーのチャット内容を分析し、具体的で実践的なフィードバックを行ってください。\n\n");

        sb.append("【コールセンター式QA 5つの評価軸】\n\n");

        sb.append("1. 共感力\n");
        sb.append("   - 相手の気持ちや状況に寄り添う言葉があるか\n");
        sb.append("   - 「お気持ちわかります」「大変でしたね」等の共感表現を使えているか\n");
        sb.append("   - 相手の感情を否定せず受け止めているか\n\n");

        sb.append("2. クッション言葉\n");
        sb.append("   - 依頼・断り・指摘の前に柔らかい前置きがあるか\n");
        sb.append("   - 「恐れ入りますが」「お手数ですが」「差し支えなければ」等を適切に使えているか\n");
        sb.append("   - 唐突な印象を与えていないか\n\n");

        sb.append("3. 結論ファースト\n");
        sb.append("   - 伝えたい要点が最初に来ているか\n");
        sb.append("   - 相手が何を求められているか一目で分かるか\n");
        sb.append("   - 長い前置きで本題が見えにくくなっていないか\n\n");

        sb.append("4. ポジティブ変換\n");
        sb.append("   - 否定的な表現を肯定的に言い換えているか\n");
        sb.append("   - 「できません」→「〜であれば可能です」のような変換ができているか\n");
        sb.append("   - 「〜しないでください」→「〜していただけると助かります」のように提案型になっているか\n\n");

        sb.append("5. 傾聴姿勢\n");
        sb.append("   - 相手の発言を確認・復唱しているか（オウム返し）\n");
        sb.append("   - 質問で相手の意図を深掘りしているか\n");
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
     * 通常チャットモード用のシステムプロンプトを構築する
     * コールセンター式コミュニケーションコーチとして応答する
     */
    public String buildCoachPrompt() {
        StringBuilder sb = new StringBuilder();

        sb.append("あなたはコールセンター出身のコミュニケーションコーチです。\n");
        sb.append("コールセンターで培った「伝わる話し方」のプロとして、ユーザーの質問や相談に答えてください。\n\n");

        sb.append("あなたの強み:\n");
        sb.append("- 共感を示しながら相手の本当のニーズを引き出す傾聴力\n");
        sb.append("- クッション言葉を使った柔らかい表現で、相手を不快にさせない伝え方\n");
        sb.append("- ネガティブな内容もポジティブに変換して伝えるスキル\n");
        sb.append("- 結論から先に話し、分かりやすく要点をまとめる力\n\n");

        sb.append("回答する際は、上記のスキルを自然に活用しながら、ユーザーにとって分かりやすく丁寧に答えてください。\n");
        sb.append("コミュニケーションに関する質問には、コールセンターでの実践的な経験に基づいたアドバイスを提供してください。");

        return sb.toString();
    }
}

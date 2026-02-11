package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

class SystemPromptBuilderTest {

    private final SystemPromptBuilder builder = new SystemPromptBuilder();

    @Nested
    @DisplayName("buildFeedbackPrompt - フィードバックモードのプロンプト構築")
    class BuildFeedbackPromptTest {

        @Test
        @DisplayName("5つのQA評価軸（共感力、クッション言葉、結論ファースト、ポジティブ変換、傾聴姿勢）が含まれる")
        void shouldContainAllFiveQaCriteria() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", "自己紹介です", "丁寧", "真面目", "スキル向上", "伝わらない", "具体的に");

            assertThat(prompt).contains("共感力");
            assertThat(prompt).contains("クッション言葉");
            assertThat(prompt).contains("結論ファースト");
            assertThat(prompt).contains("ポジティブ変換");
            assertThat(prompt).contains("傾聴姿勢");
        }

        @Test
        @DisplayName("コールセンター/QAに関連するキーワードが含まれる")
        void shouldContainCallCenterContext() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("コールセンター");
        }

        @Test
        @DisplayName("UserProfile情報がプロンプトに正しく埋め込まれる")
        void shouldEmbedUserProfileInfo() {
            String prompt = builder.buildFeedbackPrompt(
                    "山田花子", "元気な人です", "カジュアル", "明るい,社交的", "リーダーシップ", "人前で話すのが苦手", "優しく");

            assertThat(prompt).contains("山田花子");
            assertThat(prompt).contains("元気な人です");
            assertThat(prompt).contains("カジュアル");
            assertThat(prompt).contains("明るい,社交的");
            assertThat(prompt).contains("リーダーシップ");
            assertThat(prompt).contains("人前で話すのが苦手");
            assertThat(prompt).contains("優しく");
        }

        @Test
        @DisplayName("null/空のプロフィール項目はスキップされる")
        void shouldSkipNullOrEmptyProfileFields() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, "", null, "目標あり", null, "");

            assertThat(prompt).contains("テスト太郎");
            assertThat(prompt).contains("目標あり");
            assertThat(prompt).doesNotContain("自己紹介:");
            assertThat(prompt).doesNotContain("コミュニケーションスタイル:");
            assertThat(prompt).doesNotContain("性格特性:");
            assertThat(prompt).doesNotContain("懸念事項:");
            assertThat(prompt).doesNotContain("フィードバックスタイル:");
        }

        @Test
        @DisplayName("各評価軸に具体的な説明が含まれる")
        void shouldContainDetailedCriteriaDescriptions() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            // 各軸について具体的な評価ポイントが記載されていることを確認
            assertThat(prompt).containsPattern("共感力.*相手の気持ち|相手の気持ち.*共感力");
            assertThat(prompt).containsPattern("ポジティブ変換.*否定|否定.*ポジティブ変換");
        }
    }

    @Nested
    @DisplayName("buildCoachPrompt - 通常チャットモードのプロンプト構築")
    class BuildCoachPromptTest {

        @Test
        @DisplayName("コールセンター式コーチとしてのプロンプトが構築される")
        void shouldBuildCoachPrompt() {
            String prompt = builder.buildCoachPrompt();

            assertThat(prompt).contains("コールセンター");
            assertThat(prompt).contains("コミュニケーション");
        }

        @Test
        @DisplayName("5つのQA評価軸に関連するキーワードが含まれる")
        void shouldContainQaCriteriaKeywords() {
            String prompt = builder.buildCoachPrompt();

            assertThat(prompt).contains("共感");
            assertThat(prompt).contains("クッション言葉");
            assertThat(prompt).contains("ポジティブ");
        }
    }
}

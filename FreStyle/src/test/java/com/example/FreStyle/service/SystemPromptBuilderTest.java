package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

class SystemPromptBuilderTest {

    private final SystemPromptBuilder builder = new SystemPromptBuilder();

    @Nested
    @DisplayName("buildFeedbackPrompt - ビジネスコミュニケーション評価軸")
    class BuildFeedbackPromptTest {

        @Test
        @DisplayName("5つのビジネス評価軸（論理的構成力、配慮表現、要約力、提案力、質問・傾聴力）が含まれる")
        void shouldContainAllFiveBusinessCriteria() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", "自己紹介です", "丁寧", "真面目", "スキル向上", "伝わらない", "具体的に");

            assertThat(prompt).contains("論理的構成力");
            assertThat(prompt).contains("配慮表現");
            assertThat(prompt).contains("要約力");
            assertThat(prompt).contains("提案力");
            assertThat(prompt).contains("質問・傾聴力");
        }

        @Test
        @DisplayName("ビジネスコミュニケーションの文脈が含まれる")
        void shouldContainBusinessContext() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("ビジネス");
            assertThat(prompt).doesNotContain("コールセンター");
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

            assertThat(prompt).contains("PREP");
            assertThat(prompt).contains("結論");
            assertThat(prompt).contains("解決策");
        }
    }

    @Nested
    @DisplayName("buildCoachPrompt - ビジネスコミュニケーションコーチのプロンプト構築")
    class BuildCoachPromptTest {

        @Test
        @DisplayName("ビジネスコミュニケーションコーチとしてのプロンプトが構築される")
        void shouldBuildBusinessCoachPrompt() {
            String prompt = builder.buildCoachPrompt();

            assertThat(prompt).contains("ビジネスコミュニケーション");
            assertThat(prompt).doesNotContain("コールセンター");
        }

        @Test
        @DisplayName("5つのビジネス評価軸に関連するキーワードが含まれる")
        void shouldContainBusinessCriteriaKeywords() {
            String prompt = builder.buildCoachPrompt();

            assertThat(prompt).contains("論理的");
            assertThat(prompt).contains("配慮");
            assertThat(prompt).contains("提案");
        }
    }
}

package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

class SystemPromptBuilderTest {

    private final SystemPromptBuilder builder = new SystemPromptBuilder();

    @Nested
    @DisplayName("buildRephrasePrompt - 言い換え提案プロンプト")
    class BuildRephrasePromptTest {

        @Test
        @DisplayName("3パターン（フォーマル版/ソフト版/簡潔版）の指示が含まれる")
        void shouldContainThreeRephrasePatterns() {
            String prompt = builder.buildRephrasePrompt(null);

            assertThat(prompt).contains("フォーマル版");
            assertThat(prompt).contains("ソフト版");
            assertThat(prompt).contains("簡潔版");
        }

        @Test
        @DisplayName("JSON形式での出力指示が含まれる")
        void shouldContainJsonOutputInstruction() {
            String prompt = builder.buildRephrasePrompt(null);

            assertThat(prompt).contains("JSON");
        }

        @Test
        @DisplayName("シーンが指定された場合はシーンの文脈が含まれる")
        void shouldContainSceneContextWhenSpecified() {
            String prompt = builder.buildRephrasePrompt("meeting");

            assertThat(prompt).contains("会議");
        }

        @Test
        @DisplayName("シーンがnullの場合でもプロンプトが構築される")
        void shouldBuildPromptWithoutScene() {
            String prompt = builder.buildRephrasePrompt(null);

            assertThat(prompt).contains("言い換え");
            assertThat(prompt).doesNotContain("シーン");
        }

        @Test
        @DisplayName("質問型パターンの指示が含まれる")
        void shouldContainQuestioningPattern() {
            String prompt = builder.buildRephrasePrompt(null);

            assertThat(prompt).contains("質問型");
        }

        @Test
        @DisplayName("提案型パターンの指示が含まれる")
        void shouldContainProposalPattern() {
            String prompt = builder.buildRephrasePrompt(null);

            assertThat(prompt).contains("提案型");
        }

        @Test
        @DisplayName("JSON出力にquestioningフィールドが指定される")
        void shouldContainQuestioningJsonField() {
            String prompt = builder.buildRephrasePrompt(null);

            assertThat(prompt).contains("\"questioning\"");
        }

        @Test
        @DisplayName("JSON出力にproposalフィールドが指定される")
        void shouldContainProposalJsonField() {
            String prompt = builder.buildRephrasePrompt(null);

            assertThat(prompt).contains("\"proposal\"");
        }
    }

    @Nested
    @DisplayName("buildPracticePrompt - 練習モード用プロンプト構築")
    class BuildPracticePromptTest {

        @Test
        @DisplayName("シナリオ名と相手役名がプロンプトに含まれる")
        void shouldContainScenarioAndRoleName() {
            String prompt = builder.buildPracticePrompt(
                    "障害報告対応", "怒っている顧客（SIer企業のPM）", "intermediate",
                    "本番環境で障害が発生し、顧客から緊急連絡が入った状況です。");

            assertThat(prompt).contains("障害報告対応");
            assertThat(prompt).contains("怒っている顧客（SIer企業のPM）");
        }

        @Test
        @DisplayName("難易度に応じた指示が含まれる")
        void shouldContainDifficultyInstruction() {
            String beginnerPrompt = builder.buildPracticePrompt(
                    "テスト", "テスト役", "beginner", "テスト状況");

            String advancedPrompt = builder.buildPracticePrompt(
                    "テスト", "テスト役", "advanced", "テスト状況");

            assertThat(beginnerPrompt).contains("初級");
            assertThat(advancedPrompt).contains("上級");
        }

        @Test
        @DisplayName("シナリオの状況説明が含まれる")
        void shouldContainScenarioContext() {
            String prompt = builder.buildPracticePrompt(
                    "テスト", "テスト役", "intermediate",
                    "デプロイ直後に本番障害が発生した状況です。");

            assertThat(prompt).contains("デプロイ直後に本番障害が発生した状況です。");
        }

        @Test
        @DisplayName("ロールプレイの指示が含まれる")
        void shouldContainRolePlayInstruction() {
            String prompt = builder.buildPracticePrompt(
                    "テスト", "テスト役", "intermediate", "テスト状況");

            assertThat(prompt).contains("ロールプレイ");
        }

        @Test
        @DisplayName("練習終了時のスコア出力指示が含まれる")
        void shouldContainScoreOutputInstructionOnPracticeEnd() {
            String prompt = builder.buildPracticePrompt(
                    "テスト", "テスト役", "intermediate", "テスト状況");

            assertThat(prompt).contains("練習終了");
            assertThat(prompt).contains("スコア");
        }

        @Test
        @DisplayName("練習スコアのJSON形式が指定される")
        void shouldContainPracticeScoreJsonFormat() {
            String prompt = builder.buildPracticePrompt(
                    "テスト", "テスト役", "intermediate", "テスト状況");

            assertThat(prompt).contains("```json");
            assertThat(prompt).contains("\"scores\"");
        }

        @Test
        @DisplayName("練習スコアの評価軸がシナリオに応じる旨の説明が含まれる")
        void shouldContainScenarioSpecificScoreCriteria() {
            String prompt = builder.buildPracticePrompt(
                    "障害報告対応", "顧客", "intermediate", "状況");

            assertThat(prompt).contains("評価");
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

        @Test
        @DisplayName("コーチプロンプトにIT新卒向け文脈が含まれる")
        void shouldContainItJuniorContextInCoachPrompt() {
            String prompt = builder.buildCoachPrompt();

            assertThat(prompt).contains("新卒");
            assertThat(prompt).contains("エンジニア");
        }
    }
}

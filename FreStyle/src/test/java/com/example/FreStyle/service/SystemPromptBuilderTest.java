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

        @Test
        @DisplayName("JSONスコア出力形式の指示が含まれる")
        void shouldContainJsonScoreOutputInstruction() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("```json");
            assertThat(prompt).contains("\"scores\"");
            assertThat(prompt).contains("\"axis\"");
            assertThat(prompt).contains("\"score\"");
            assertThat(prompt).contains("\"comment\"");
        }
    }

    @Nested
    @DisplayName("buildFeedbackPromptWithScene - シーン別フィードバックプロンプト")
    class BuildFeedbackPromptWithSceneTest {

        @Test
        @DisplayName("会議シーンの追加評価観点が含まれる")
        void shouldContainMeetingSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "meeting", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("会議");
            assertThat(prompt).contains("発言のタイミング");
            assertThat(prompt).contains("議論の建設性");
            assertThat(prompt).contains("ファシリテーション");
            // 基本5軸も含まれること
            assertThat(prompt).contains("論理的構成力");
            assertThat(prompt).contains("配慮表現");
        }

        @Test
        @DisplayName("1on1シーンの追加評価観点が含まれる")
        void shouldContainOneOnOneSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "one_on_one", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("1on1");
            assertThat(prompt).contains("心理的安全性");
            assertThat(prompt).contains("フィードバックの具体性");
            assertThat(prompt).contains("傾聴の深さ");
        }

        @Test
        @DisplayName("メールシーンの追加評価観点が含まれる")
        void shouldContainEmailSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "email", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("メール");
            assertThat(prompt).contains("件名の明確さ");
            assertThat(prompt).contains("構成の読みやすさ");
            assertThat(prompt).contains("アクション明示");
        }

        @Test
        @DisplayName("プレゼンシーンの追加評価観点が含まれる")
        void shouldContainPresentationSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "presentation", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("プレゼン");
            assertThat(prompt).contains("ストーリー構成");
            assertThat(prompt).contains("聞き手への配慮");
            assertThat(prompt).contains("質疑応答力");
        }

        @Test
        @DisplayName("商談シーンの追加評価観点が含まれる")
        void shouldContainNegotiationSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "negotiation", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("商談");
            assertThat(prompt).contains("ニーズヒアリング");
            assertThat(prompt).contains("価値提案");
            assertThat(prompt).contains("クロージング");
        }

        @Test
        @DisplayName("コードレビューシーンの評価観点が含まれる")
        void shouldContainCodeReviewSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "code_review", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("コードレビュー");
            assertThat(prompt).contains("指摘の具体性");
            assertThat(prompt).contains("代替案の提示");
            assertThat(prompt).contains("相手への配慮");
        }

        @Test
        @DisplayName("障害対応シーンの評価観点が含まれる")
        void shouldContainIncidentSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "incident", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("障害対応");
            assertThat(prompt).contains("状況報告の正確さ");
            assertThat(prompt).contains("エスカレーション判断");
            assertThat(prompt).contains("事後報告の構成");
        }

        @Test
        @DisplayName("日報シーンの評価観点が含まれる")
        void shouldContainDailyReportSceneCriteria() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "daily_report", "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("日報");
            assertThat(prompt).contains("成果の定量化");
            assertThat(prompt).contains("課題の明確化");
            assertThat(prompt).contains("ネクストアクション");
        }

        @Test
        @DisplayName("シーンがnullの場合は基本5軸のみのプロンプトが返される")
        void shouldReturnBasicPromptWhenSceneIsNull() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    null, "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("論理的構成力");
            assertThat(prompt).contains("配慮表現");
            assertThat(prompt).contains("要約力");
            assertThat(prompt).contains("提案力");
            assertThat(prompt).contains("質問・傾聴力");
            // シーン別の追加観点は含まれない
            assertThat(prompt).doesNotContain("【シーン別追加評価観点】");
        }

        @Test
        @DisplayName("UserProfile情報がシーン別プロンプトにも正しく埋め込まれる")
        void shouldEmbedUserProfileInScenePrompt() {
            String prompt = builder.buildFeedbackPromptWithScene(
                    "meeting", "山田花子", "元気な人です", "カジュアル", "明るい", "リーダーシップ", "緊張する", "優しく");

            assertThat(prompt).contains("山田花子");
            assertThat(prompt).contains("元気な人です");
            assertThat(prompt).contains("カジュアル");
            assertThat(prompt).contains("リーダーシップ");
        }
    }

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
    }

    @Nested
    @DisplayName("IT新卒向け評価軸サブ基準")
    class ItJuniorSubCriteriaTest {

        @Test
        @DisplayName("報連相の構造化サブ基準が含まれる")
        void shouldContainHouRenSouSubCriteria() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("報連相");
        }

        @Test
        @DisplayName("敬語の正確さサブ基準が含まれる")
        void shouldContainKeigoSubCriteria() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("敬語の正確さ");
        }

        @Test
        @DisplayName("技術説明の平易化サブ基準が含まれる")
        void shouldContainTechnicalExplanationSubCriteria() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("技術説明の平易化");
        }

        @Test
        @DisplayName("エスカレーション判断力サブ基準が含まれる")
        void shouldContainEscalationSubCriteria() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("エスカレーション判断力");
        }

        @Test
        @DisplayName("要件確認の網羅性サブ基準が含まれる")
        void shouldContainRequirementConfirmationSubCriteria() {
            String prompt = builder.buildFeedbackPrompt(
                    "テスト太郎", null, null, null, null, null, null);

            assertThat(prompt).contains("要件確認の網羅性");
        }

        @Test
        @DisplayName("コーチプロンプトにIT新卒向け文脈が含まれる")
        void shouldContainItJuniorContextInCoachPrompt() {
            String prompt = builder.buildCoachPrompt();

            assertThat(prompt).contains("新卒");
            assertThat(prompt).contains("エンジニア");
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

package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

class ScoreCardServiceTest {

    private final ScoreCardService service = new ScoreCardService(null);

    @Nested
    @DisplayName("parseScoresFromResponse - AI応答からスコアJSON抽出")
    class ParseScoresTest {

        @Test
        @DisplayName("正常なJSON形式のスコアを抽出できる")
        void shouldParseValidScoreJson() {
            String aiResponse = "フィードバック内容...\n```json\n" +
                    "{\"scores\":[" +
                    "{\"axis\":\"論理的構成力\",\"score\":8,\"comment\":\"良い\"}," +
                    "{\"axis\":\"配慮表現\",\"score\":6,\"comment\":\"改善余地\"}," +
                    "{\"axis\":\"要約力\",\"score\":7,\"comment\":\"まあまあ\"}," +
                    "{\"axis\":\"提案力\",\"score\":5,\"comment\":\"要改善\"}," +
                    "{\"axis\":\"質問・傾聴力\",\"score\":9,\"comment\":\"優秀\"}" +
                    "]}\n```";

            List<ScoreCardService.AxisScore> scores = service.parseScoresFromResponse(aiResponse);

            assertThat(scores).hasSize(5);
            assertThat(scores.get(0).getAxis()).isEqualTo("論理的構成力");
            assertThat(scores.get(0).getScore()).isEqualTo(8);
            assertThat(scores.get(0).getComment()).isEqualTo("良い");
        }

        @Test
        @DisplayName("JSON部分がない場合は空リストを返す")
        void shouldReturnEmptyListWhenNoJson() {
            String aiResponse = "フィードバックのみで、スコアJSON無し";

            List<ScoreCardService.AxisScore> scores = service.parseScoresFromResponse(aiResponse);

            assertThat(scores).isEmpty();
        }

        @Test
        @DisplayName("不正なJSON形式の場合は空リストを返す")
        void shouldReturnEmptyListForInvalidJson() {
            String aiResponse = "```json\n{invalid json}\n```";

            List<ScoreCardService.AxisScore> scores = service.parseScoresFromResponse(aiResponse);

            assertThat(scores).isEmpty();
        }
    }

    @Nested
    @DisplayName("calculateOverallScore - 総合スコア計算")
    class CalculateOverallScoreTest {

        @Test
        @DisplayName("各軸スコアの平均値が計算される")
        void shouldCalculateAverageScore() {
            List<ScoreCardService.AxisScore> scores = List.of(
                    new ScoreCardService.AxisScore("論理的構成力", 8, "良い"),
                    new ScoreCardService.AxisScore("配慮表現", 6, "改善余地"),
                    new ScoreCardService.AxisScore("要約力", 7, "まあまあ"),
                    new ScoreCardService.AxisScore("提案力", 5, "要改善"),
                    new ScoreCardService.AxisScore("質問・傾聴力", 9, "優秀")
            );

            double overall = service.calculateOverallScore(scores);

            assertThat(overall).isEqualTo(7.0);
        }

        @Test
        @DisplayName("空リストの場合は0.0を返す")
        void shouldReturnZeroForEmptyList() {
            double overall = service.calculateOverallScore(List.of());

            assertThat(overall).isEqualTo(0.0);
        }
    }
}

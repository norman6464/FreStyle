package com.example.FreStyle.service;

import java.util.ArrayList;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import org.springframework.stereotype.Service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import lombok.extern.slf4j.Slf4j;

/**
 * スコアカードのユーティリティサービス
 *
 * <p>役割:</p>
 * <ul>
 *   <li>AI応答テキストからスコアJSONを抽出・パースする</li>
 *   <li>スコアの平均値を計算する</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層のユーティリティサービス</li>
 *   <li>UseCaseから呼び出されるパースロジック</li>
 * </ul>
 */
@Service
@Slf4j
public class ScoreCardService {

    private final ObjectMapper objectMapper = new ObjectMapper();

    /**
     * AI応答テキストからスコアJSONを抽出してパースする
     */
    public List<AxisScore> parseScoresFromResponse(String aiResponse) {
        List<AxisScore> scores = new ArrayList<>();

        try {
            // ```json ... ``` ブロックからJSONを抽出
            Pattern pattern = Pattern.compile("```json\\s*\\n?(.*?)\\n?```", Pattern.DOTALL);
            Matcher matcher = pattern.matcher(aiResponse);

            if (!matcher.find()) {
                log.debug("AI応答にJSONブロックが見つかりません");
                return scores;
            }

            String jsonStr = matcher.group(1).trim();
            JsonNode root = objectMapper.readTree(jsonStr);
            JsonNode scoresNode = root.path("scores");

            if (scoresNode.isMissingNode() || !scoresNode.isArray()) {
                log.debug("JSONにscores配列が見つかりません");
                return scores;
            }

            for (JsonNode scoreNode : scoresNode) {
                String axis = scoreNode.path("axis").asText();
                int score = scoreNode.path("score").asInt();
                String comment = scoreNode.path("comment").asText();
                scores.add(new AxisScore(axis, score, comment));
            }

        } catch (Exception e) {
            log.warn("スコアJSONのパースに失敗しました: {}", e.getMessage());
        }

        return scores;
    }

    /**
     * 各軸スコアの平均値を計算する
     */
    public double calculateOverallScore(List<AxisScore> scores) {
        if (scores == null || scores.isEmpty()) {
            return 0.0;
        }
        return scores.stream()
                .mapToInt(AxisScore::getScore)
                .average()
                .orElse(0.0);
    }

    /**
     * 評価軸スコアの内部表現
     */
    public static class AxisScore {
        private final String axis;
        private final int score;
        private final String comment;

        public AxisScore(String axis, int score, String comment) {
            this.axis = axis;
            this.score = score;
            this.comment = comment;
        }

        public String getAxis() {
            return axis;
        }

        public int getScore() {
            return score;
        }

        public String getComment() {
            return comment;
        }
    }
}

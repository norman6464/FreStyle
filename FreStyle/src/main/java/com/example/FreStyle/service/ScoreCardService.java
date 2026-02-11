package com.example.FreStyle.service;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.dto.ScoreHistoryDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Service
@RequiredArgsConstructor
@Slf4j
public class ScoreCardService {

    private final CommunicationScoreRepository communicationScoreRepository;
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
     * スコアをDBに保存する
     */
    public void saveScores(Integer sessionId, Integer userId, List<AxisScore> scores, String scene) {
        AiChatSession session = new AiChatSession();
        session.setId(sessionId);

        User user = new User();
        user.setId(userId);

        for (AxisScore axisScore : scores) {
            CommunicationScore entity = new CommunicationScore();
            entity.setSession(session);
            entity.setUser(user);
            entity.setAxisName(axisScore.getAxis());
            entity.setScore(axisScore.getScore());
            entity.setComment(axisScore.getComment());
            entity.setScene(scene);
            communicationScoreRepository.save(entity);
        }

        log.info("スコア保存完了 - sessionId: {}, 軸数: {}", sessionId, scores.size());
    }

    /**
     * セッションIDからスコアカードを取得する
     */
    public ScoreCardDto getScoreCard(Integer sessionId) {
        List<CommunicationScore> entities = communicationScoreRepository.findBySessionId(sessionId);

        List<ScoreCardDto.AxisScoreDto> scoreDtos = entities.stream()
                .map(e -> new ScoreCardDto.AxisScoreDto(e.getAxisName(), e.getScore(), e.getComment()))
                .toList();

        List<AxisScore> axisScores = entities.stream()
                .map(e -> new AxisScore(e.getAxisName(), e.getScore(), e.getComment()))
                .toList();

        double overall = calculateOverallScore(axisScores);

        return new ScoreCardDto(sessionId, scoreDtos, overall);
    }

    /**
     * ユーザーのスコア履歴を取得する
     */
    public List<CommunicationScore> getScoreHistory(Integer userId) {
        return communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);
    }

    /**
     * ユーザーのスコア履歴をセッション単位でグループ化して取得する
     */
    public List<ScoreHistoryDto> getScoreHistoryGrouped(Integer userId) {
        List<CommunicationScore> allScores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);

        // セッションIDでグループ化（順序保持）
        Map<Integer, List<CommunicationScore>> grouped = new LinkedHashMap<>();
        for (CommunicationScore score : allScores) {
            grouped.computeIfAbsent(score.getSession().getId(), k -> new ArrayList<>()).add(score);
        }

        List<ScoreHistoryDto> history = new ArrayList<>();
        for (Map.Entry<Integer, List<CommunicationScore>> entry : grouped.entrySet()) {
            List<CommunicationScore> sessionScores = entry.getValue();
            CommunicationScore first = sessionScores.get(0);

            List<ScoreCardDto.AxisScoreDto> scoreDtos = sessionScores.stream()
                    .map(s -> new ScoreCardDto.AxisScoreDto(s.getAxisName(), s.getScore(), s.getComment()))
                    .toList();

            List<AxisScore> axisScores = sessionScores.stream()
                    .map(s -> new AxisScore(s.getAxisName(), s.getScore(), s.getComment()))
                    .toList();

            double overall = calculateOverallScore(axisScores);
            String title = first.getSession().getTitle();

            history.add(new ScoreHistoryDto(
                    entry.getKey(), title, overall, scoreDtos, first.getCreatedAt()));
        }

        return history;
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

package com.example.FreStyle.mapper;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Component;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.dto.ScoreHistoryDto;
import com.example.FreStyle.entity.CommunicationScore;

/**
 * ScoreCardのマッピングクラス
 *
 * <p>役割:</p>
 * <ul>
 *   <li>CommunicationScoreエンティティ ⇔ ScoreCardDto/ScoreHistoryDtoの変換</li>
 *   <li>プレゼンテーション層とドメイン層の境界を明確化</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>プレゼンテーション層とアプリケーション層の間のマッピング層</li>
 *   <li>DTOとEntityの変換ロジックを一箇所に集約</li>
 * </ul>
 */
@Component
public class ScoreCardMapper {

    /**
     * CommunicationScoreエンティティリストからScoreCardDtoへ変換
     *
     * @param sessionId セッションID
     * @param scores CommunicationScoreエンティティのリスト
     * @return ScoreCardDto（APIレスポンス用）
     * @throws IllegalArgumentException scoresがnullの場合
     */
    public ScoreCardDto toScoreCardDto(Integer sessionId, List<CommunicationScore> scores) {
        if (scores == null) {
            throw new IllegalArgumentException("scoresがnullです");
        }

        List<ScoreCardDto.AxisScoreDto> scoreDtos = scores.stream()
                .map(this::toAxisScoreDto)
                .toList();

        double overallScore = calculateOverallScore(scores);

        return new ScoreCardDto(sessionId, scoreDtos, overallScore);
    }

    /**
     * CommunicationScoreエンティティリストをセッション単位でグループ化し、ScoreHistoryDtoリストへ変換
     *
     * @param scores CommunicationScoreエンティティのリスト（createdAt降順）
     * @return ScoreHistoryDtoのリスト
     * @throws IllegalArgumentException scoresがnullの場合
     */
    public List<ScoreHistoryDto> toScoreHistoryDtoList(List<CommunicationScore> scores) {
        if (scores == null) {
            throw new IllegalArgumentException("scoresがnullです");
        }

        // セッションIDでグループ化（順序保持）
        Map<Integer, List<CommunicationScore>> grouped = scores.stream()
                .collect(Collectors.groupingBy(
                        s -> s.getSession().getId(),
                        LinkedHashMap::new,
                        Collectors.toList()));

        return grouped.entrySet().stream()
                .map(entry -> {
                    List<CommunicationScore> sessionScores = entry.getValue();
                    CommunicationScore first = sessionScores.getFirst();
                    List<ScoreCardDto.AxisScoreDto> scoreDtos = sessionScores.stream()
                            .map(this::toAxisScoreDto)
                            .toList();
                    double overallScore = calculateOverallScore(sessionScores);
                    return new ScoreHistoryDto(
                            entry.getKey(), first.getSession().getTitle(),
                            overallScore, scoreDtos, first.getCreatedAt());
                })
                .toList();
    }

    /**
     * CommunicationScoreエンティティからAxisScoreDtoへ変換
     */
    private ScoreCardDto.AxisScoreDto toAxisScoreDto(CommunicationScore score) {
        return new ScoreCardDto.AxisScoreDto(
                score.getAxisName(),
                score.getScore(),
                score.getComment()
        );
    }

    /**
     * CommunicationScoreリストから平均スコアを計算
     */
    private double calculateOverallScore(List<CommunicationScore> scores) {
        if (scores.isEmpty()) {
            return 0.0;
        }
        return scores.stream()
                .mapToInt(CommunicationScore::getScore)
                .average()
                .orElse(0.0);
    }
}

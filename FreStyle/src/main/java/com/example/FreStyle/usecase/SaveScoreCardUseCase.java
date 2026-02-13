package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.service.ScoreCardService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/**
 * スコアカード保存ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>AI応答テキストからスコアを抽出する</li>
 *   <li>抽出したスコアをデータベースに保存する</li>
 *   <li>保存結果をScoreCardDtoとして返却する</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層のユースケース</li>
 *   <li>プレゼンテーション層（Controller）から呼び出される</li>
 *   <li>ドメイン層（Repository）に依存</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
@Slf4j
public class SaveScoreCardUseCase {

    private final CommunicationScoreRepository communicationScoreRepository;
    private final ScoreCardService scoreCardService;

    /**
     * AI応答からスコアを抽出・保存し、ScoreCardDtoを返す
     *
     * @param sessionId セッションID
     * @param userId ユーザーID
     * @param aiResponse AI応答テキスト
     * @param scene フィードバックシーン（nullable）
     * @return ScoreCardDto（スコアが抽出できない場合はnull）
     */
    @Transactional
    public ScoreCardDto execute(Integer sessionId, Integer userId, String aiResponse, String scene) {
        List<ScoreCardService.AxisScore> scores = scoreCardService.parseScoresFromResponse(aiResponse);

        if (scores.isEmpty()) {
            return null;
        }

        AiChatSession session = new AiChatSession();
        session.setId(sessionId);

        User user = new User();
        user.setId(userId);

        for (ScoreCardService.AxisScore axisScore : scores) {
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

        double overallScore = scoreCardService.calculateOverallScore(scores);

        List<ScoreCardDto.AxisScoreDto> scoreDtos = scores.stream()
                .map(s -> new ScoreCardDto.AxisScoreDto(s.getAxis(), s.getScore(), s.getComment()))
                .toList();

        return new ScoreCardDto(sessionId, scoreDtos, overallScore);
    }
}

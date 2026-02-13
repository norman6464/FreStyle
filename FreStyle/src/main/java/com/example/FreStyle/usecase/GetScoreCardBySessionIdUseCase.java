package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ScoreCardDto;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.mapper.ScoreCardMapper;
import com.example.FreStyle.repository.CommunicationScoreRepository;

import lombok.RequiredArgsConstructor;

/**
 * セッションIDによるスコアカード取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたセッションIDに紐づくスコアカードを取得する</li>
 *   <li>CommunicationScoreエンティティをScoreCardDtoに変換して返却する</li>
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
public class GetScoreCardBySessionIdUseCase {

    private final CommunicationScoreRepository communicationScoreRepository;
    private final ScoreCardMapper mapper;

    /**
     * セッションIDでスコアカードを取得する
     *
     * @param sessionId セッションID
     * @return ScoreCardDto（スコアカード情報）
     */
    @Transactional(readOnly = true)
    public ScoreCardDto execute(Integer sessionId) {
        List<CommunicationScore> scores = communicationScoreRepository.findBySessionId(sessionId);
        return mapper.toScoreCardDto(sessionId, scores);
    }
}

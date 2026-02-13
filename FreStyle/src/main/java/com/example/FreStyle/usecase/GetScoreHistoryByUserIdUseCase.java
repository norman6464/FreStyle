package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ScoreHistoryDto;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.mapper.ScoreCardMapper;
import com.example.FreStyle.repository.CommunicationScoreRepository;

import lombok.RequiredArgsConstructor;

/**
 * ユーザーIDによるスコア履歴取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたユーザーIDに紐づくスコア履歴をセッション単位で取得する</li>
 *   <li>CommunicationScoreエンティティをScoreHistoryDtoに変換して返却する</li>
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
public class GetScoreHistoryByUserIdUseCase {

    private final CommunicationScoreRepository communicationScoreRepository;
    private final ScoreCardMapper mapper;

    /**
     * ユーザーIDでスコア履歴を取得する
     *
     * @param userId ユーザーID
     * @return ScoreHistoryDtoのリスト（セッション単位でグループ化済み）
     */
    @Transactional(readOnly = true)
    public List<ScoreHistoryDto> execute(Integer userId) {
        List<CommunicationScore> scores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);
        return mapper.toScoreHistoryDtoList(scores);
    }
}

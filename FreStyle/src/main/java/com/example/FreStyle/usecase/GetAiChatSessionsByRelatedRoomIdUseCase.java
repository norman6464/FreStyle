package com.example.FreStyle.usecase;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;

import lombok.RequiredArgsConstructor;

/**
 * 関連ルーム別AI Chatセッション一覧取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたチャットルームに関連するAI Chatセッションを取得</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>関連ルームIDが一致するセッションを返却</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class GetAiChatSessionsByRelatedRoomIdUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final AiChatSessionMapper mapper;

    /**
     * 関連ルームIDでセッション一覧を取得
     *
     * @param roomId チャットルームID
     * @return セッションDTOのリスト
     */
    @Transactional(readOnly = true)
    public List<AiChatSessionDto> execute(Integer roomId) {
        List<AiChatSession> sessions = aiChatSessionRepository.findByRelatedRoomId(roomId);
        return sessions.stream()
                .map(mapper::toDto)
                .collect(Collectors.toList());
    }
}

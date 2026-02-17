package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;

import lombok.RequiredArgsConstructor;

/**
 * ユーザー別AI Chatセッション一覧取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたユーザーIDのすべてのAI Chatセッションを取得</li>
 *   <li>作成日時の降順でソート</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>セッションは作成日時の新しい順で返却</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class GetAiChatSessionsByUserIdUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final AiChatSessionMapper mapper;

    /**
     * 指定ユーザーのセッション一覧を取得
     *
     * @param userId ユーザーID
     * @return セッションDTOのリスト（作成日時降順）
     */
    @Transactional(readOnly = true)
    public List<AiChatSessionDto> execute(Integer userId) {
        List<AiChatSession> sessions = aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId);
        return sessions.stream()
                .map(mapper::toDto)
                .toList();
    }
}

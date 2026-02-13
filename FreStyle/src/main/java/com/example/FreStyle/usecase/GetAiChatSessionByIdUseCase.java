package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;

import lombok.RequiredArgsConstructor;

/**
 * AI Chatセッション取得ユースケース（権限チェック付き）
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたセッションIDとユーザーIDでセッションを取得</li>
 *   <li>ユーザーの権限チェック（自分のセッションのみ取得可能）</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>セッションIDとユーザーIDが一致しない場合はエラー</li>
 *   <li>セッションが存在しない場合はエラー</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class GetAiChatSessionByIdUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final AiChatSessionMapper mapper;

    /**
     * セッションIDとユーザーIDでセッションを取得
     *
     * @param sessionId セッションID
     * @param userId ユーザーID
     * @return セッションDTO
     * @throws RuntimeException セッションが見つからない、または権限がない場合
     */
    @Transactional(readOnly = true)
    public AiChatSessionDto execute(Integer sessionId, Integer userId) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new RuntimeException(
                    "セッションが見つからないか、アクセス権限がありません: sessionId=" + sessionId + ", userId=" + userId
                ));
        return mapper.toDto(session);
    }
}

package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;

import lombok.RequiredArgsConstructor;

/**
 * AI Chatセッションタイトル更新ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたセッションのタイトルを更新</li>
 *   <li>ユーザーの権限チェック（自分のセッションのみ更新可能）</li>
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
public class UpdateAiChatSessionTitleUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final AiChatSessionMapper mapper;

    /**
     * セッションタイトルを更新
     *
     * @param sessionId セッションID
     * @param userId ユーザーID
     * @param newTitle 新しいタイトル
     * @return 更新されたセッションDTO
     * @throws ResourceNotFoundException セッションが見つからない、または権限がない場合
     */
    @Transactional
    public AiChatSessionDto execute(Integer sessionId, Integer userId, String newTitle) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new ResourceNotFoundException(
                    "セッションが見つからないか、アクセス権限がありません: sessionId=" + sessionId + ", userId=" + userId
                ));

        session.setTitle(newTitle);
        AiChatSession saved = aiChatSessionRepository.save(session);
        return mapper.toDto(saved);
    }
}

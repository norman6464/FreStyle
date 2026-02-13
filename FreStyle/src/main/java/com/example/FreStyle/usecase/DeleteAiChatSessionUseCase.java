package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.repository.AiChatSessionRepository;

import lombok.RequiredArgsConstructor;

/**
 * AI Chatセッション削除ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたセッションを削除</li>
 *   <li>ユーザーの権限チェック（自分のセッションのみ削除可能）</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>セッションIDとユーザーIDが一致しない場合はエラー</li>
 *   <li>セッションが存在しない場合はエラー</li>
 *   <li>削除時、関連するメッセージも自動削除される（カスケード削除）</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class DeleteAiChatSessionUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;

    /**
     * セッションを削除
     *
     * @param sessionId セッションID
     * @param userId ユーザーID
     * @throws RuntimeException セッションが見つからない、または権限がない場合
     */
    @Transactional
    public void execute(Integer sessionId, Integer userId) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new RuntimeException(
                    "セッションが見つからないか、アクセス権限がありません: sessionId=" + sessionId + ", userId=" + userId
                ));

        aiChatSessionRepository.delete(session);
    }
}

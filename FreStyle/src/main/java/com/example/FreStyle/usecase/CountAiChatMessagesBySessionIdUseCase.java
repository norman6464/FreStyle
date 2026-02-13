package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.repository.AiChatMessageRepository;

import lombok.RequiredArgsConstructor;

/**
 * セッション別AI Chatメッセージ数取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたセッションIDのメッセージ数をカウント</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>ユーザーメッセージとアシスタントメッセージの両方をカウント</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class CountAiChatMessagesBySessionIdUseCase {

    private final AiChatMessageRepository aiChatMessageRepository;

    /**
     * 指定セッションのメッセージ数を取得
     *
     * @param sessionId セッションID
     * @return メッセージ数
     */
    @Transactional(readOnly = true)
    public Long execute(Integer sessionId) {
        return aiChatMessageRepository.countBySessionId(sessionId);
    }
}

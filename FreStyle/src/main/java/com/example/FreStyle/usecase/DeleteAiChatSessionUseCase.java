package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;

import lombok.RequiredArgsConstructor;

/**
 * AI Chatセッション削除ユースケース
 * DynamoDBのメッセージを明示的に削除してからRDBのセッションを削除する
 */
@Service
@RequiredArgsConstructor
public class DeleteAiChatSessionUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    /**
     * セッションを削除（DynamoDBメッセージ + RDBセッション）
     */
    @Transactional
    public void execute(Integer sessionId, Integer userId) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(sessionId, userId)
                .orElseThrow(() -> new ResourceNotFoundException(
                    "セッションが見つからないか、アクセス権限がありません: sessionId=" + sessionId + ", userId=" + userId
                ));

        // DynamoDBからメッセージを削除
        aiChatMessageDynamoRepository.deleteBySessionId(sessionId);

        // RDBからセッションを削除
        aiChatSessionRepository.delete(session);
    }
}

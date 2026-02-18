package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;

/**
 * AI Chatメッセージ追加ユースケース
 */
@Service
@RequiredArgsConstructor
public class AddAiChatMessageUseCase {

    private final AiChatMessageDynamoRepository aiChatMessageDynamoRepository;
    private final AiChatSessionRepository aiChatSessionRepository;
    private final UserRepository userRepository;

    /**
     * メッセージを追加（汎用）
     */
    public AiChatMessageResponseDto execute(Integer sessionId, Integer userId, String role, String content) {
        aiChatSessionRepository.findById(sessionId)
                .orElseThrow(() -> new ResourceNotFoundException("セッションが見つかりません: ID=" + sessionId));

        userRepository.findById(userId)
                .orElseThrow(() -> new ResourceNotFoundException("ユーザーが見つかりません: ID=" + userId));

        return aiChatMessageDynamoRepository.save(sessionId, userId, role, content);
    }

    /**
     * ユーザーメッセージを追加
     */
    public AiChatMessageResponseDto executeUserMessage(Integer sessionId, Integer userId, String content) {
        return execute(sessionId, userId, "user", content);
    }

    /**
     * アシスタントメッセージを追加
     */
    public AiChatMessageResponseDto executeAssistantMessage(Integer sessionId, Integer userId, String content) {
        return execute(sessionId, userId, "assistant", content);
    }
}

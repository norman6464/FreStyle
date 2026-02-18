package com.example.FreStyle.service;

import java.util.List;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AiChatMessageService {

    private final AiChatMessageDynamoRepository aiChatMessageDynamoRepository;
    private final AiChatSessionRepository aiChatSessionRepository;
    private final UserRepository userRepository;

    /**
     * 新しいメッセージを追加
     */
    public AiChatMessageResponseDto addMessage(Integer sessionId, Integer userId, String role, String content) {
        // セッションとユーザーの存在確認はRDBで行う
        aiChatSessionRepository.findById(sessionId)
                .orElseThrow(() -> new RuntimeException("セッションが見つかりません: " + sessionId));

        userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("ユーザーが見つかりません: " + userId));

        return aiChatMessageDynamoRepository.save(sessionId, userId, role, content);
    }

    /**
     * ユーザーメッセージを追加
     */
    public AiChatMessageResponseDto addUserMessage(Integer sessionId, Integer userId, String content) {
        return addMessage(sessionId, userId, "user", content);
    }

    /**
     * アシスタントメッセージを追加
     */
    public AiChatMessageResponseDto addAssistantMessage(Integer sessionId, Integer userId, String content) {
        return addMessage(sessionId, userId, "assistant", content);
    }

    /**
     * 指定セッションのメッセージ一覧を取得
     */
    public List<AiChatMessageResponseDto> getMessagesBySessionId(Integer sessionId) {
        return aiChatMessageDynamoRepository.findBySessionId(sessionId);
    }

    /**
     * 指定ユーザーの全メッセージを取得
     */
    public List<AiChatMessageResponseDto> getMessagesByUserId(Integer userId) {
        return aiChatMessageDynamoRepository.findByUserId(userId);
    }

    /**
     * 指定セッションのメッセージ数を取得
     */
    public Long countMessagesBySessionId(Integer sessionId) {
        return aiChatMessageDynamoRepository.countBySessionId(sessionId);
    }
}

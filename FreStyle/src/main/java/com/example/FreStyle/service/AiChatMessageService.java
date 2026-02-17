package com.example.FreStyle.service;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.entity.AiChatMessage.Role;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.AiChatMessageRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AiChatMessageService {

    private final AiChatMessageRepository aiChatMessageRepository;
    private final AiChatSessionRepository aiChatSessionRepository;
    private final UserRepository userRepository;

    /**
     * 新しいメッセージを追加
     */
    @Transactional
    public AiChatMessageResponseDto addMessage(Integer sessionId, Integer userId, Role role, String content) {
        AiChatSession session = aiChatSessionRepository.findById(sessionId)
                .orElseThrow(() -> new RuntimeException("セッションが見つかりません: " + sessionId));

        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("ユーザーが見つかりません: " + userId));

        AiChatMessage message = new AiChatMessage();
        message.setSession(session);
        message.setUser(user);
        message.setRole(role);
        message.setContent(content);

        AiChatMessage saved = aiChatMessageRepository.save(message);
        return toDto(saved);
    }

    /**
     * ユーザーメッセージを追加
     */
    @Transactional
    public AiChatMessageResponseDto addUserMessage(Integer sessionId, Integer userId, String content) {
        return addMessage(sessionId, userId, Role.user, content);
    }

    /**
     * アシスタントメッセージを追加
     */
    @Transactional
    public AiChatMessageResponseDto addAssistantMessage(Integer sessionId, Integer userId, String content) {
        return addMessage(sessionId, userId, Role.assistant, content);
    }

    /**
     * 指定セッションのメッセージ一覧を取得
     */
    @Transactional(readOnly = true)
    public List<AiChatMessageResponseDto> getMessagesBySessionId(Integer sessionId) {
        List<AiChatMessage> messages = aiChatMessageRepository.findBySessionIdOrderByCreatedAtAsc(sessionId);
        return messages.stream()
                .map(this::toDto)
                .toList();
    }

    /**
     * 指定ユーザーの全メッセージを取得
     */
    @Transactional(readOnly = true)
    public List<AiChatMessageResponseDto> getMessagesByUserId(Integer userId) {
        List<AiChatMessage> messages = aiChatMessageRepository.findByUserIdOrderByCreatedAtAsc(userId);
        return messages.stream()
                .map(this::toDto)
                .toList();
    }

    /**
     * 指定セッションのメッセージ数を取得
     */
    @Transactional(readOnly = true)
    public Long countMessagesBySessionId(Integer sessionId) {
        return aiChatMessageRepository.countBySessionId(sessionId);
    }

    /**
     * EntityからDTOへ変換
     */
    private AiChatMessageResponseDto toDto(AiChatMessage message) {
        return new AiChatMessageResponseDto(
                message.getId(),
                message.getSession().getId(),
                message.getUser().getId(),
                message.getRole().name(),
                message.getContent(),
                message.getCreatedAt()
        );
    }

}

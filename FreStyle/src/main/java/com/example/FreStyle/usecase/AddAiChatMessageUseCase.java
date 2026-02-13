package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.entity.AiChatMessage.Role;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.mapper.AiChatMessageMapper;
import com.example.FreStyle.repository.AiChatMessageRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;

/**
 * AI Chatメッセージ追加ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>新しいAI Chatメッセージを追加</li>
 *   <li>ユーザーメッセージ、アシスタントメッセージの追加をサポート</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>セッションIDとユーザーIDが有効である必要がある</li>
 *   <li>ロール（user/assistant）を指定してメッセージを追加</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class AddAiChatMessageUseCase {

    private final AiChatMessageRepository aiChatMessageRepository;
    private final AiChatSessionRepository aiChatSessionRepository;
    private final UserRepository userRepository;
    private final AiChatMessageMapper mapper;

    /**
     * メッセージを追加（汎用）
     *
     * @param sessionId セッションID
     * @param userId ユーザーID
     * @param role ロール（user/assistant）
     * @param content メッセージ内容
     * @return 追加されたメッセージDTO
     * @throws ResourceNotFoundException セッションまたはユーザーが見つからない場合
     */
    @Transactional
    public AiChatMessageResponseDto execute(Integer sessionId, Integer userId, Role role, String content) {
        AiChatSession session = aiChatSessionRepository.findById(sessionId)
                .orElseThrow(() -> new ResourceNotFoundException("セッションが見つかりません: ID=" + sessionId));

        User user = userRepository.findById(userId)
                .orElseThrow(() -> new ResourceNotFoundException("ユーザーが見つかりません: ID=" + userId));

        AiChatMessage message = new AiChatMessage();
        message.setSession(session);
        message.setUser(user);
        message.setRole(role);
        message.setContent(content);

        AiChatMessage saved = aiChatMessageRepository.save(message);
        return mapper.toDto(saved);
    }

    /**
     * ユーザーメッセージを追加
     *
     * @param sessionId セッションID
     * @param userId ユーザーID
     * @param content メッセージ内容
     * @return 追加されたメッセージDTO
     * @throws ResourceNotFoundException セッションまたはユーザーが見つからない場合
     */
    @Transactional
    public AiChatMessageResponseDto executeUserMessage(Integer sessionId, Integer userId, String content) {
        return execute(sessionId, userId, Role.user, content);
    }

    /**
     * アシスタントメッセージを追加
     *
     * @param sessionId セッションID
     * @param userId ユーザーID
     * @param content メッセージ内容
     * @return 追加されたメッセージDTO
     * @throws ResourceNotFoundException セッションまたはユーザーが見つからない場合
     */
    @Transactional
    public AiChatMessageResponseDto executeAssistantMessage(Integer sessionId, Integer userId, String content) {
        return execute(sessionId, userId, Role.assistant, content);
    }
}

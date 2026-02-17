package com.example.FreStyle.service;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatMessage;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ChatMessageRepository;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Service
@Slf4j
@RequiredArgsConstructor
public class ChatMessageService {

    private final ChatMessageRepository chatMessageRepository;
    private final UserService userService;

    /**
     * 指定ルームのチャット履歴取得（作成日時昇順）
     * 現在のユーザーIDを受け取り、isSenderフラグを設定
     */
    @Transactional(readOnly = true)
    public List<ChatMessageDto> getMessagesByRoom(ChatRoom room, Integer currentUserId) {
        List<ChatMessage> messages = chatMessageRepository.findByRoomOrderByCreatedAtAsc(room);
        return messages.stream()
                .map(msg -> toDto(msg))
                .toList();
    }

    /**
     * 新しいメッセージを保存
     */
    @Transactional
    public ChatMessageDto addMessage(ChatRoom room, Integer senderId, String content) {
        log.debug("addMessage - roomId: {}, senderId: {}", room.getId(), senderId);

        User sender = userService.findUserById(senderId);

        ChatMessage message = new ChatMessage();
        message.setRoom(room);
        message.setSender(sender);
        message.setContent(content);

        ChatMessage saved = chatMessageRepository.save(message);
        log.debug("メッセージ保存成功 - messageId: {}", saved.getId());

        return toDto(saved);
    }

    /**
     * メッセージを更新（ID指定）
     */
    @Transactional
    public ChatMessageDto updateMessage(Integer messageId, String newContent) {
        ChatMessage message = chatMessageRepository.findById(messageId)
                .orElseThrow(() -> new RuntimeException("メッセージが見つかりません。"));

        message.setContent(newContent);
        ChatMessage updated = chatMessageRepository.save(message);
        return toDto(updated);
    }

    /**
     * メッセージ削除
     */
    @Transactional
    public void deleteMessage(Integer messageId) {
        if (!chatMessageRepository.existsById(messageId)) {
            throw new RuntimeException("メッセージが見つかりません。");
        }
        chatMessageRepository.deleteById(messageId);
    }

    /**
     * ChatMessage → ChatMessageDto 変換（isSenderフラグ付き）
     */
    private ChatMessageDto toDto(ChatMessage message) {
        return new ChatMessageDto(
                message.getId(),
                message.getRoom().getId(),
                message.getSender().getId(),
                message.getSender().getName(),
                message.getContent(),
                message.getCreatedAt(),
                message.getUpdatedAt()
        );
    }
}

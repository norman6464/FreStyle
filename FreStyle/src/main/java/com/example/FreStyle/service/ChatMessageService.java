package com.example.FreStyle.service;

import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ChatMessageDynamoRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Service
@Slf4j
@RequiredArgsConstructor
public class ChatMessageService {

    private final ChatMessageDynamoRepository chatMessageDynamoRepository;
    private final UserRepository userRepository;

    /**
     * 指定ルームのチャット履歴取得（作成日時昇順）
     * DynamoDB取得後、RDBからsenderNameをバッチ解決
     */
    public List<ChatMessageDto> getMessagesByRoom(Integer roomId, Integer currentUserId) {
        List<ChatMessageDto> messages = chatMessageDynamoRepository.findByRoomId(roomId);

        if (messages.isEmpty()) {
            return messages;
        }

        // senderIdの一覧を取得してバッチでUser情報を取得
        Set<Integer> senderIds = messages.stream()
                .map(ChatMessageDto::senderId)
                .collect(Collectors.toSet());
        Map<Integer, String> senderNameMap = userRepository.findAllById(senderIds).stream()
                .collect(Collectors.toMap(User::getId, User::getName));

        // senderNameを解決して新しいDTOを返す
        return messages.stream()
                .map(msg -> new ChatMessageDto(
                        msg.id(),
                        msg.roomId(),
                        msg.senderId(),
                        senderNameMap.getOrDefault(msg.senderId(), null),
                        msg.content(),
                        msg.createdAt()
                ))
                .toList();
    }

    /**
     * 新しいメッセージを保存
     */
    public ChatMessageDto addMessage(Integer roomId, Integer senderId, String content) {
        log.debug("addMessage - roomId: {}, senderId: {}", roomId, senderId);

        ChatMessageDto saved = chatMessageDynamoRepository.save(roomId, senderId, content);
        log.debug("メッセージ保存成功 - messageId: {}", saved.id());

        // senderNameを解決
        User sender = userRepository.findById(senderId).orElse(null);
        String senderName = sender != null ? sender.getName() : null;

        return new ChatMessageDto(
                saved.id(),
                saved.roomId(),
                saved.senderId(),
                senderName,
                saved.content(),
                saved.createdAt()
        );
    }

    /**
     * メッセージ削除（roomId + createdAt で特定）
     */
    public void deleteMessage(Integer roomId, Long createdAt) {
        chatMessageDynamoRepository.deleteByRoomIdAndCreatedAt(roomId, createdAt);
    }
}

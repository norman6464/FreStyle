package com.example.FreStyle.repository;

import java.util.List;
import java.util.Map;

import com.example.FreStyle.dto.ChatMessageDto;

/**
 * ユーザー間チャットメッセージのDynamoDBリポジトリインターフェース
 */
public interface ChatMessageDynamoRepository {

    ChatMessageDto save(Integer roomId, Integer senderId, String content);

    List<ChatMessageDto> findByRoomId(Integer roomId);

    void deleteByRoomIdAndCreatedAt(Integer roomId, Long createdAt);

    ChatMessageDto findLatestByRoomId(Integer roomId);

    Map<Integer, ChatMessageDto> findLatestByRoomIds(List<Integer> roomIds);
}

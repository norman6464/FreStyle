package com.example.FreStyle.repository;

import java.util.List;

import com.example.FreStyle.dto.AiChatMessageResponseDto;

/**
 * AIチャットメッセージのDynamoDBリポジトリインターフェース
 */
public interface AiChatMessageDynamoRepository {

    AiChatMessageResponseDto save(Integer sessionId, Integer userId, String role, String content);

    List<AiChatMessageResponseDto> findBySessionId(Integer sessionId);

    List<AiChatMessageResponseDto> findByUserId(Integer userId);

    Long countBySessionId(Integer sessionId);

    void deleteBySessionId(Integer sessionId);
}

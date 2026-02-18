package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;

import lombok.RequiredArgsConstructor;

/**
 * セッション別AI Chatメッセージ一覧取得ユースケース
 */
@Service
@RequiredArgsConstructor
public class GetAiChatMessagesBySessionIdUseCase {

    private final AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    /**
     * 指定セッションのメッセージ一覧を取得
     */
    public List<AiChatMessageResponseDto> execute(Integer sessionId) {
        return aiChatMessageDynamoRepository.findBySessionId(sessionId);
    }
}

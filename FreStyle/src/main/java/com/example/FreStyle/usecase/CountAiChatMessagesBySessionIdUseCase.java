package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.repository.AiChatMessageDynamoRepository;

import lombok.RequiredArgsConstructor;

/**
 * セッション別AI Chatメッセージ数取得ユースケース
 */
@Service
@RequiredArgsConstructor
public class CountAiChatMessagesBySessionIdUseCase {

    private final AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    /**
     * 指定セッションのメッセージ数を取得
     */
    public Long execute(Integer sessionId) {
        return aiChatMessageDynamoRepository.countBySessionId(sessionId);
    }
}

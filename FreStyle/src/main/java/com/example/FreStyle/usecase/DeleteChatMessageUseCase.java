package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.service.ChatMessageService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class DeleteChatMessageUseCase {

    private final ChatMessageService chatMessageService;

    public void execute(Integer roomId, Long createdAt) {
        chatMessageService.deleteMessage(roomId, createdAt);
    }
}

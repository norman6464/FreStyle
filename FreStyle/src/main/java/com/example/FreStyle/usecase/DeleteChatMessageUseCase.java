package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.service.ChatMessageService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class DeleteChatMessageUseCase {

    private final ChatMessageService chatMessageService;

    @Transactional
    public void execute(Integer messageId) {
        chatMessageService.deleteMessage(messageId);
    }
}

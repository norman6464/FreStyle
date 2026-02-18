package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetChatHistoryUseCase {

    private final UserIdentityService userIdentityService;
    private final ChatMessageService chatMessageService;

    public List<ChatMessageDto> execute(String sub, Integer roomId) {
        User user = userIdentityService.findUserBySub(sub);
        return chatMessageService.getMessagesByRoom(roomId, user.getId());
    }
}

package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class CreateOrGetChatRoomUseCase {

    private final UserIdentityService userIdentityService;
    private final ChatService chatService;

    @Transactional
    public Integer execute(String sub, Integer targetUserId) {
        User user = userIdentityService.findUserBySub(sub);
        return chatService.createOrGetRoom(user.getId(), targetUserId);
    }
}

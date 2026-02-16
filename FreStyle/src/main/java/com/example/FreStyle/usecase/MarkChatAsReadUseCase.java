package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UnreadCountService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class MarkChatAsReadUseCase {

    private final UserIdentityService userIdentityService;
    private final UnreadCountService unreadCountService;

    @Transactional
    public void execute(String sub, Integer roomId) {
        User user = userIdentityService.findUserBySub(sub);
        unreadCountService.resetUnreadCount(user.getId(), roomId);
    }
}

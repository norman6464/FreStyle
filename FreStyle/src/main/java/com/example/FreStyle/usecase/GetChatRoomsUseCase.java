package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetChatRoomsUseCase {

    private final UserIdentityService userIdentityService;
    private final ChatService chatService;

    @Transactional(readOnly = true)
    public List<ChatUserDto> execute(String sub, String query) {
        User user = userIdentityService.findUserBySub(sub);
        return chatService.findChatUsers(user.getId(), query);
    }
}

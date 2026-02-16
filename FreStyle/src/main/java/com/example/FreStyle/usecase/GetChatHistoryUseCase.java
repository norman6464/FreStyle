package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetChatHistoryUseCase {

    private final UserIdentityService userIdentityService;
    private final ChatRoomService chatRoomService;
    private final ChatMessageService chatMessageService;

    @Transactional(readOnly = true)
    public List<ChatMessageDto> execute(String sub, Integer roomId) {
        User user = userIdentityService.findUserBySub(sub);
        ChatRoom chatRoom = chatRoomService.findChatRoomById(roomId);
        return chatMessageService.getMessagesByRoom(chatRoom, user.getId());
    }
}

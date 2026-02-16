package com.example.FreStyle.usecase;

import java.util.Optional;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.UnreadCountService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class SendChatMessageUseCase {

    private final ChatRoomService chatRoomService;
    private final ChatMessageService chatMessageService;
    private final UnreadCountService unreadCountService;
    private final RoomMemberRepository roomMemberRepository;

    public record Result(ChatMessageDto message, Integer partnerId) {}

    @Transactional
    public Result execute(Integer senderId, Integer roomId, String content) {
        ChatRoom room = chatRoomService.findChatRoomById(roomId);
        ChatMessageDto saved = chatMessageService.addMessage(room, senderId, content);

        Optional<User> partnerOpt = roomMemberRepository.findPartnerByRoomIdAndUserId(roomId, senderId);
        Integer partnerId = null;
        if (partnerOpt.isPresent()) {
            partnerId = partnerOpt.get().getId();
            unreadCountService.incrementUnreadCount(partnerId, roomId);
        }

        return new Result(saved, partnerId);
    }
}

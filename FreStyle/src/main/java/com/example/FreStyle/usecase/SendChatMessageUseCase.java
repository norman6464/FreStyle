package com.example.FreStyle.usecase;

import java.util.Optional;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.UnreadCountService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class SendChatMessageUseCase {

    private final ChatMessageService chatMessageService;
    private final UnreadCountService unreadCountService;
    private final RoomMemberRepository roomMemberRepository;

    public record Result(ChatMessageDto message, Integer partnerId) {}

    public Result execute(Integer senderId, Integer roomId, String content) {
        ChatMessageDto saved = chatMessageService.addMessage(roomId, senderId, content);

        Optional<User> partnerOpt = roomMemberRepository.findPartnerByRoomIdAndUserId(roomId, senderId);
        Integer partnerId = null;
        if (partnerOpt.isPresent()) {
            partnerId = partnerOpt.get().getId();
            unreadCountService.incrementUnreadCount(partnerId, roomId);
        }

        return new Result(saved, partnerId);
    }
}

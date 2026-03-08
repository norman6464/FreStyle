package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetChatHistoryUseCase {

    private final UserIdentityService userIdentityService;
    private final ChatMessageService chatMessageService;
    private final RoomMemberRepository roomMemberRepository;

    public List<ChatMessageDto> execute(String sub, Integer roomId) {
        User user = userIdentityService.findUserBySub(sub);

        // ルームメンバーシップチェック（IDOR防止）
        if (!roomMemberRepository.existsByRoom_IdAndUser_Id(roomId, user.getId())) {
            throw new SecurityException("このルームのメッセージを閲覧する権限がありません");
        }

        return chatMessageService.getMessagesByRoom(roomId, user.getId());
    }
}

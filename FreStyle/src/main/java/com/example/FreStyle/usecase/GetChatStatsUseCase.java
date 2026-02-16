package com.example.FreStyle.usecase;

import java.util.HashMap;
import java.util.Map;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.RoomMemberService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetChatStatsUseCase {

    private final UserIdentityService userIdentityService;
    private final RoomMemberService roomMemberService;

    @Transactional(readOnly = true)
    public Map<String, Object> execute(String sub) {
        User user = userIdentityService.findUserBySub(sub);
        Long chatPartnerCount = roomMemberService.countChatPartners(user.getId());

        Map<String, Object> stats = new HashMap<>();
        stats.put("chatPartnerCount", chatPartnerCount);
        stats.put("email", user.getEmail());
        stats.put("username", user.getName());
        return stats;
    }
}

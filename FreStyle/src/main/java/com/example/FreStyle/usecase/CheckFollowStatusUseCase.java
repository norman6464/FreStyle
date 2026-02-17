package com.example.FreStyle.usecase;

import java.util.LinkedHashMap;
import java.util.Map;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.repository.FriendshipRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class CheckFollowStatusUseCase {

    private final FriendshipRepository friendshipRepository;

    @Transactional(readOnly = true)
    public Map<String, Object> execute(Integer userId, Integer targetUserId) {
        boolean isFollowing = friendshipRepository.existsByFollowerIdAndFollowingId(userId, targetUserId);
        boolean isFollowedBy = friendshipRepository.existsByFollowerIdAndFollowingId(targetUserId, userId);

        Map<String, Object> result = new LinkedHashMap<>();
        result.put("isFollowing", isFollowing);
        result.put("isFollowedBy", isFollowedBy);
        result.put("isMutual", isFollowing && isFollowedBy);
        return result;
    }
}

package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FollowStatusDto;
import com.example.FreStyle.repository.FriendshipRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class CheckFollowStatusUseCase {

    private final FriendshipRepository friendshipRepository;

    @Transactional(readOnly = true)
    public FollowStatusDto execute(Integer userId, Integer targetUserId) {
        boolean isFollowing = friendshipRepository.existsByFollowerIdAndFollowingId(userId, targetUserId);
        boolean isFollowedBy = friendshipRepository.existsByFollowerIdAndFollowingId(targetUserId, userId);
        return new FollowStatusDto(isFollowing, isFollowedBy, isFollowing && isFollowedBy);
    }
}

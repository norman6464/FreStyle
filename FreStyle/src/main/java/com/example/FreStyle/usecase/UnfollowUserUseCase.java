package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.FriendshipRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UnfollowUserUseCase {

    private final FriendshipRepository friendshipRepository;

    @Transactional
    public void execute(Integer followerId, Integer followingId) {
        Friendship friendship = friendshipRepository
                .findByFollowerIdAndFollowingId(followerId, followingId)
                .orElseThrow(() -> new ResourceNotFoundException("フォロー関係が見つかりません"));

        friendshipRepository.delete(friendship);
    }
}

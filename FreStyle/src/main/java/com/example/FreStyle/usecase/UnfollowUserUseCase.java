package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.FriendshipRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UnfollowUserUseCase {

    private final FriendshipRepository friendshipRepository;

    @Transactional
    public void execute(User follower, Integer followingId) {
        Friendship friendship = friendshipRepository
                .findByFollowerIdAndFollowingId(follower.getId(), followingId)
                .orElseThrow(() -> new ResourceNotFoundException("フォロー関係が見つかりません"));

        friendshipRepository.delete(friendship);
    }
}

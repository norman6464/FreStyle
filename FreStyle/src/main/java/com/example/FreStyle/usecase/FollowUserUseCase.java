package com.example.FreStyle.usecase;

import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.FriendshipMapper;
import com.example.FreStyle.repository.FriendshipRepository;
import com.example.FreStyle.service.UserService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class FollowUserUseCase {

    private final FriendshipRepository friendshipRepository;
    private final FriendshipMapper friendshipMapper;
    private final UserService userService;

    @Transactional
    public FriendshipDto execute(User follower, Integer followingId) {
        if (follower.getId().equals(followingId)) {
            throw new IllegalArgumentException("自分自身をフォローすることはできません");
        }

        if (friendshipRepository.existsByFollowerIdAndFollowingId(follower.getId(), followingId)) {
            throw new IllegalArgumentException("既にフォローしています");
        }

        User following = userService.findUserById(followingId);

        Friendship friendship = new Friendship();
        friendship.setFollower(follower);
        friendship.setFollowing(following);

        Friendship saved;
        try {
            saved = friendshipRepository.save(friendship);
        } catch (DataIntegrityViolationException e) {
            throw new IllegalArgumentException("既にフォローしています");
        }

        boolean mutual = friendshipRepository.existsByFollowerIdAndFollowingId(
                followingId, follower.getId());

        return friendshipMapper.toFollowingDto(saved, mutual);
    }
}

package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.FriendshipMapper;
import com.example.FreStyle.repository.FriendshipRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class FollowUserUseCase {

    private final FriendshipRepository friendshipRepository;
    private final FriendshipMapper friendshipMapper;

    @Transactional
    public FriendshipDto execute(User follower, User following) {
        if (follower.getId().equals(following.getId())) {
            throw new IllegalArgumentException("自分自身をフォローすることはできません");
        }

        if (friendshipRepository.existsByFollowerIdAndFollowingId(follower.getId(), following.getId())) {
            throw new IllegalStateException("既にフォローしています");
        }

        Friendship friendship = new Friendship();
        friendship.setFollower(follower);
        friendship.setFollowing(following);

        Friendship saved = friendshipRepository.save(friendship);

        boolean mutual = friendshipRepository.existsByFollowerIdAndFollowingId(
                following.getId(), follower.getId());

        return friendshipMapper.toFollowingDto(saved, mutual);
    }
}

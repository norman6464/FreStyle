package com.example.FreStyle.mapper;

import org.springframework.stereotype.Component;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;

@Component
public class FriendshipMapper {

    public FriendshipDto toFollowingDto(Friendship friendship, boolean mutual) {
        User following = friendship.getFollowing();
        return new FriendshipDto(
                friendship.getId(),
                following.getId(),
                following.getName(),
                following.getIconUrl(),
                following.getBio(),
                mutual,
                friendship.getCreatedAt() != null ? friendship.getCreatedAt().toString() : null
        );
    }

    public FriendshipDto toFollowerDto(Friendship friendship, boolean mutual) {
        User follower = friendship.getFollower();
        return new FriendshipDto(
                friendship.getId(),
                follower.getId(),
                follower.getName(),
                follower.getIconUrl(),
                follower.getBio(),
                mutual,
                friendship.getCreatedAt() != null ? friendship.getCreatedAt().toString() : null
        );
    }
}

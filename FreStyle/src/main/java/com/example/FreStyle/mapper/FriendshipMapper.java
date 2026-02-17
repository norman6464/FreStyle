package com.example.FreStyle.mapper;

import org.springframework.stereotype.Component;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;

@Component
public class FriendshipMapper {

    public FriendshipDto toFollowingDto(Friendship friendship, boolean mutual) {
        return toDto(friendship, friendship.getFollowing(), mutual);
    }

    public FriendshipDto toFollowerDto(Friendship friendship, boolean mutual) {
        return toDto(friendship, friendship.getFollower(), mutual);
    }

    private FriendshipDto toDto(Friendship friendship, User user, boolean mutual) {
        return new FriendshipDto(
                friendship.getId(),
                user.getId(),
                user.getName(),
                user.getIconUrl(),
                user.getBio(),
                mutual,
                friendship.getCreatedAt() != null ? friendship.getCreatedAt().toString() : null
        );
    }
}

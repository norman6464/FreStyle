package com.example.FreStyle.dto;

public record FollowStatusDto(
        boolean isFollowing,
        boolean isFollowedBy,
        boolean isMutual
) {
}

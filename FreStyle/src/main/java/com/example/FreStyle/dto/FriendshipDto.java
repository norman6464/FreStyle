package com.example.FreStyle.dto;

public record FriendshipDto(
        Integer id,
        Integer userId,
        String username,
        String iconUrl,
        String bio,
        boolean mutual,
        String createdAt,
        String status) {
}

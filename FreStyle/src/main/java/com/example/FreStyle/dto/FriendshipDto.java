package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class FriendshipDto {
    private Integer id;
    private Integer userId;
    private String username;
    private String iconUrl;
    private String bio;
    private boolean mutual;
    private String createdAt;
}

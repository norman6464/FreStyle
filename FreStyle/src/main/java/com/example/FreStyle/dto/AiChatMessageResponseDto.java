package com.example.FreStyle.dto;

import java.sql.Timestamp;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class AiChatMessageResponseDto {
    private Integer id;
    private Integer sessionId;
    private Integer userId;
    private String role;        // "user" または "assistant"
    private String content;
    private Timestamp createdAt;
}

package com.example.FreStyle.dto;

import java.sql.Timestamp;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class AiChatSessionDto {
    private Integer id;
    private Integer userId;
    private String title;
    private Integer relatedRoomId;
    private String scene;
    private String sessionType;
    private Integer scenarioId;
    private Timestamp createdAt;
    private Timestamp updatedAt;
}

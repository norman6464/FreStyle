package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class NotificationDto {
    private Integer id;
    private String type;
    private String title;
    private String message;
    private Boolean isRead;
    private Integer relatedId;
    private String createdAt;
}

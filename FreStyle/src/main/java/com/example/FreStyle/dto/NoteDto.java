package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class NoteDto {
    private String noteId;
    private Integer userId;
    private String title;
    private String content;
    private Boolean isPinned;
    private Long createdAt;
    private Long updatedAt;
}

package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class SessionNoteDto {
    private Integer sessionId;
    private String note;
    private String updatedAt;
}

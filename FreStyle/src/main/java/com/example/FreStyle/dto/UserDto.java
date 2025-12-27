package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class UserDto {
    private Integer id;
    private String email;
    private String name;
    private Integer roomId; // 新しく追加

    public UserDto(Integer id, String email, String name) {
        this.id = id;
        this.email = email;
        this.name = name;
    }
}

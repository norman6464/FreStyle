package com.example.FreStyle.dto;

public class UserDto {
    private Integer id;
    private String email;
    private Integer roomId; // 新しく追加

    public UserDto(Integer id, String email) {
        this.id = id;
        this.email = email;
    }

    public UserDto() {}

    public Integer getId() {
        return id;
    }

    public void setId(Integer id) {
        this.id = id;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public Integer getRoomId() {
        return roomId;
    }

    public void setRoomId(Integer roomId) {
        this.roomId = roomId;
    }
}

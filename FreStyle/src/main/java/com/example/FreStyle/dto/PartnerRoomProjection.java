package com.example.FreStyle.dto;

/**
 * 相手ユーザーIDとルームIDのペアを型安全に受け取るためのプロジェクション
 */
public interface PartnerRoomProjection {
    Integer getUserId();
    Integer getRoomId();
}

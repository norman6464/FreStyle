package com.example.FreStyle.dto;

import java.sql.Timestamp;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * チャット履歴のあるユーザー情報DTO
 * ユーザー情報 + ルーム情報 + 最終メッセージ情報を含む
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ChatUserDto {
    // ユーザー情報
    private Integer userId;
    private String email;
    private String name;
    
    // ルーム情報
    private Integer roomId;
    
    // 最終メッセージ情報
    private String lastMessage;
    private Integer lastMessageSenderId;
    private String lastMessageSenderName;
    private Timestamp lastMessageAt;
    
    // 未読数（将来の拡張用）
    private Integer unreadCount;

    // プロフィール画像URL
    private String profileImage;
    
    /**
     * 最終メッセージなしの場合のコンストラクタ
     */
    public ChatUserDto(Integer userId, String email, String name, Integer roomId) {
        this.userId = userId;
        this.email = email;
        this.name = name;
        this.roomId = roomId;
        this.unreadCount = 0;
    }
}

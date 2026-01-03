package com.example.FreStyle.service;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatMessage;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ChatMessageRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class ChatMessageService {

    private final ChatMessageRepository chatMessageRepository;
    private final UserService userService;

    /**
     * 指定ルームのチャット履歴取得（作成日時昇順）
     * 現在のユーザーIDを受け取り、isSenderフラグを設定
     */
    @Transactional(readOnly = true)
    public List<ChatMessageDto> getMessagesByRoom(ChatRoom room, Integer currentUserId) {
        List<ChatMessage> messages = chatMessageRepository.findByRoomOrderByCreatedAtAsc(room);
        return messages.stream()
                .map(msg -> toDto(msg))
                .collect(Collectors.toList());
    }

    /**
     * 新しいメッセージを保存
     */
    @Transactional
    public ChatMessageDto addMessage(ChatRoom room, Integer senderId, String content) {
        System.out.println("\n========== ChatMessageService.addMessage() 実行 ==========");
        System.out.println("📝 パラメータ:");
        System.out.println("   - room.id: " + room.getId());
        System.out.println("   - senderId: " + senderId + " (タイプ: String)");
        System.out.println("   - content: " + content);
        
        try {
            // ユーザーを Cognito sub で検索
            System.out.println("🔍 UserIdentityService.findUserBySub(\"" + senderId + "\") を実行中...");
            System.out.println("   💡 注意: senderId がユーザーID (Integer) か、Cognito sub (String) かを確認");
            // senderIdをInrtegerに変換してUserを取得する
            User sender = userService.findUserById(senderId);
            
            System.out.println("✅ ユーザー取得成功");
            System.out.println("   - sender.id: " + sender.getId());
            System.out.println("   - sender.name: " + sender.getName());
            System.out.println("   - sender.email: " + sender.getEmail());
            
            // メッセージオブジェクトを作成
            ChatMessage message = new ChatMessage();
            message.setRoom(room);
            message.setSender(sender);
            message.setContent(content);
            
            System.out.println("💾 ChatMessageRepository.save() を実行中...");
            ChatMessage saved = chatMessageRepository.save(message);
            
            System.out.println("✅ メッセージ保存成功");
            System.out.println("   - messageId: " + saved.getId());
            System.out.println("   - roomId: " + saved.getRoom().getId());
            System.out.println("   - senderId: " + saved.getSender().getId());
            System.out.println("========== ChatMessageService.addMessage() 完了 ==========\n");
            
            return toDto(saved);
            
        } catch (RuntimeException e) {
            System.out.println("❌ RuntimeException 発生");
            System.out.println("   - エラーメッセージ: " + e.getMessage());
            System.out.println("   - 原因: senderId=\"" + senderId + "\" が Cognito sub として見つかりません");
            System.out.println("   - 推測: senderId がユーザーID (Integer) なのに、Cognito sub (UUID形式) を期待しています");
            System.out.println("========== ChatMessageService.addMessage() 失敗 ==========\n");
            e.printStackTrace();
            throw e;
        } catch (Exception e) {
            System.out.println("❌ 予期しないエラー発生");
            System.out.println("   - エラータイプ: " + e.getClass().getName());
            System.out.println("   - エラーメッセージ: " + e.getMessage());
            System.out.println("========== ChatMessageService.addMessage() 失敗 ==========\n");
            e.printStackTrace();
            throw e;
        }
    }

    /**
     * メッセージを更新（ID指定）
     */
    @Transactional
    public ChatMessageDto updateMessage(Integer messageId, String newContent) {
        ChatMessage message = chatMessageRepository.findById(messageId)
                .orElseThrow(() -> new RuntimeException("メッセージが見つかりません。"));

        message.setContent(newContent);
        ChatMessage updated = chatMessageRepository.save(message);
        return toDto(updated);
    }

    /**
     * メッセージ削除
     */
    @Transactional
    public void deleteMessage(Integer messageId) {
        if (!chatMessageRepository.existsById(messageId)) {
            throw new RuntimeException("メッセージが見つかりません。");
        }
        chatMessageRepository.deleteById(messageId);
    }

    /**
     * ChatMessage → ChatMessageDto 変換（isSenderフラグ付き）
     */
    private ChatMessageDto toDto(ChatMessage message) {
        return new ChatMessageDto(
                message.getId(),
                message.getRoom().getId(),
                message.getSender().getId(),
                message.getSender().getName(),
                message.getContent(),
                message.getCreatedAt(),
                message.getUpdatedAt()
        );
    }
}

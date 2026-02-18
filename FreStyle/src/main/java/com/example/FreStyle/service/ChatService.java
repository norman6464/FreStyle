package com.example.FreStyle.service;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.dto.ChatUserDto;
import lombok.extern.slf4j.Slf4j;
import com.example.FreStyle.dto.PartnerRoomProjection;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.RoomMember;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ChatMessageDynamoRepository;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.repository.UserRepository;
import lombok.RequiredArgsConstructor;

// ChatRoomServiceとRoomMemberServiceクラス二つとも関与しているときはこちらのクラスを使う
@Service
@Slf4j
@RequiredArgsConstructor
public class ChatService {
    private final ChatRoomRepository chatRoomRepository;
    private final RoomMemberRepository roomMemberRepository;
    private final UserRepository userRepository;
    private final ChatMessageDynamoRepository chatMessageDynamoRepository;
    private final UnreadCountService unreadCountService;

    
    // チャットルームの作成かすでに存在をしていた場合はそのままチャット画面のページへ移動をする
    @Transactional
    public Integer createOrGetRoom(Integer myUserId, Integer targetUserId) {
      
      Integer existingRoomId = chatRoomRepository.findRoomIdByUserIds(myUserId, targetUserId);
      if (existingRoomId != null) {
        return existingRoomId;
      }
      
      ChatRoom newRoom = new ChatRoom();
      chatRoomRepository.save(newRoom);
      
      RoomMember myMember = new RoomMember();
      myMember.setRoom(newRoom);
      myMember.setUser(userRepository.findById(myUserId)
              .orElseThrow(()-> new IllegalStateException("ユーザーが存在しません。")));
      
      RoomMember targetMember = new RoomMember();
      targetMember.setRoom(newRoom);
      targetMember.setUser(userRepository.findById(targetUserId)
                  .orElseThrow(() -> new IllegalStateException("相手ユーザーが存在しません。")));
      
      roomMemberRepository.saveAll(List.of(myMember, targetMember));
                  
      return newRoom.getId();
      
    }

    /**
     * チャット履歴のあるユーザー一覧を取得（最終メッセージ情報付き）
     * @param myUserId 自分のユーザーID
     * @param query 検索クエリ（名前またはメールで検索、nullの場合は全件）
     * @return チャット履歴のあるユーザー一覧
     */
    @Transactional(readOnly = true)
    public List<ChatUserDto> findChatUsers(Integer myUserId, String query) {
        log.debug("findChatUsers 開始 - myUserId: {}, query: {}", myUserId, query);

        // 1. 自分が参加しているルームと相手ユーザーIDのペアを取得
        List<PartnerRoomProjection> partnerData = roomMemberRepository.findPartnerUserIdAndRoomIdByUserId(myUserId);
        log.debug("取得したルーム数: {}", partnerData.size());

        if (partnerData.isEmpty()) {
            return new ArrayList<>();
        }

        // 2. 相手ユーザーIDとルームIDのマップを作成
        Map<Integer, Integer> userIdToRoomId = new HashMap<>();
        List<Integer> roomIds = new ArrayList<>();
        List<Integer> partnerUserIds = new ArrayList<>();

        for (PartnerRoomProjection row : partnerData) {
            Integer partnerId = row.getUserId();
            Integer roomId = row.getRoomId();
            userIdToRoomId.put(partnerId, roomId);
            roomIds.add(roomId);
            partnerUserIds.add(partnerId);
        }
        
        // 3. 相手ユーザー情報を取得
        List<User> partners = userRepository.findAllById(partnerUserIds);
        Map<Integer, User> userMap = partners.stream()
                .collect(Collectors.toMap(User::getId, u -> u));
        
        // 4. 各ルームの最新メッセージをDynamoDBから一括取得
        Map<Integer, ChatMessageDto> roomToLatestMessage = chatMessageDynamoRepository.findLatestByRoomIds(roomIds);

        // 5. 未読数を一括取得
        Map<Integer, Integer> unreadCounts = unreadCountService.getUnreadCountsByUserAndRooms(myUserId, roomIds);

        // 6. DTOを構築
        List<ChatUserDto> result = new ArrayList<>();

        for (Integer partnerId : partnerUserIds) {
            User partner = userMap.get(partnerId);
            if (partner == null) continue;

            // 検索クエリがある場合はフィルタリング
            if (query != null && !query.isEmpty()) {
                String lowerQuery = query.toLowerCase();
                boolean matchesName = partner.getName() != null &&
                        partner.getName().toLowerCase().contains(lowerQuery);
                boolean matchesEmail = partner.getEmail() != null &&
                        partner.getEmail().toLowerCase().contains(lowerQuery);
                if (!matchesName && !matchesEmail) {
                    continue;
                }
            }

            Integer roomId = userIdToRoomId.get(partnerId);
            ChatMessageDto latestMsg = roomToLatestMessage.get(roomId);

            ChatUserDto dto = new ChatUserDto();
            dto.setUserId(partner.getId());
            dto.setEmail(partner.getEmail());
            dto.setName(partner.getName());
            dto.setRoomId(roomId);
            dto.setProfileImage(partner.getIconUrl());
            dto.setUnreadCount(unreadCounts.getOrDefault(roomId, 0));

            if (latestMsg != null) {
                dto.setLastMessage(latestMsg.content());
                dto.setLastMessageSenderId(latestMsg.senderId());
                // senderNameをuserMapから解決
                User senderUser = userMap.get(latestMsg.senderId());
                dto.setLastMessageSenderName(senderUser != null ? senderUser.getName() : null);
                dto.setLastMessageAt(latestMsg.createdAt());
            }

            result.add(dto);
        }
        
        // 6. 最終メッセージ日時で降順ソート（新しい順）
        result.sort((a, b) -> {
            if (a.getLastMessageAt() == null && b.getLastMessageAt() == null) return 0;
            if (a.getLastMessageAt() == null) return 1;
            if (b.getLastMessageAt() == null) return -1;
            return b.getLastMessageAt().compareTo(a.getLastMessageAt());
        });
        
        log.debug("findChatUsers 完了 - 結果数: {}", result.size());
        return result;
    }
}

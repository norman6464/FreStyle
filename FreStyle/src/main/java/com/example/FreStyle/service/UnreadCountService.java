package com.example.FreStyle.service;

import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.UnreadCount;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.UnreadCountRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UnreadCountService {

    private final UnreadCountRepository unreadCountRepository;
    private final UserService userService;
    private final ChatRoomService chatRoomService;

    /**
     * 指定ユーザー・ルームの未読カウントを+1する
     * レコードが存在しない場合は新規作成してカウント1にする
     */
    @Transactional
    public void incrementUnreadCount(Integer userId, Integer roomId) {
        User user = userService.findUserById(userId);
        ChatRoom room = chatRoomService.findChatRoomById(roomId);

        UnreadCount unreadCount = unreadCountRepository.findByUserAndRoom(user, room)
                .orElseGet(() -> {
                    UnreadCount newCount = new UnreadCount();
                    newCount.setUser(user);
                    newCount.setRoom(room);
                    newCount.setCount(0);
                    return newCount;
                });

        unreadCount.setCount(unreadCount.getCount() + 1);
        unreadCountRepository.save(unreadCount);
    }

    /**
     * 指定ユーザー・ルームの未読カウントを0にリセットする
     * レコードが存在しない場合は何もしない
     */
    @Transactional
    public void resetUnreadCount(Integer userId, Integer roomId) {
        User user = userService.findUserById(userId);
        ChatRoom room = chatRoomService.findChatRoomById(roomId);

        unreadCountRepository.findByUserAndRoom(user, room)
                .ifPresent(uc -> {
                    uc.setCount(0);
                    unreadCountRepository.save(uc);
                });
    }

    /**
     * ユーザーの複数ルームに対する未読カウントを一括取得する
     * @return roomId → count のMap
     */
    @Transactional(readOnly = true)
    public Map<Integer, Integer> getUnreadCountsByUserAndRooms(Integer userId, List<Integer> roomIds) {
        if (roomIds.isEmpty()) {
            return Collections.emptyMap();
        }

        List<UnreadCount> counts = unreadCountRepository.findByUserIdAndRoomIds(userId, roomIds);
        return counts.stream()
                .collect(Collectors.toMap(
                        uc -> uc.getRoom().getId(),
                        UnreadCount::getCount
                ));
    }
}

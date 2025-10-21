package com.example.FreStyle.repository;

import com.example.FreStyle.entity.ChatRoom;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface ChatRoomRepository extends JpaRepository<ChatRoom, Integer> {
    // 必要に応じてカスタムメソッドを追加可能
}

package com.example.FreStyle.repository;

import com.example.FreStyle.entity.RoomMember;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface RoomMemberRepository extends JpaRepository<RoomMember, Integer> {
    // ユニーク制約（room_id と user_id）に基づく検索メソッドの例
    boolean existsByRoomIdAndUserId(Integer roomId, Integer userId);
}

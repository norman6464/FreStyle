package com.example.FreStyle.repository;

import com.example.FreStyle.entity.UnreadCount;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface UnreadCountRepository extends JpaRepository<UnreadCount, Integer> {

    // user と room による検索
    Optional<UnreadCount> findByUserAndRoom(User user, ChatRoom room);

    // ユーザーIDと複数ルームIDによる一括取得（N+1回避）
    @Query("SELECT uc FROM UnreadCount uc WHERE uc.user.id = :userId AND uc.room.id IN :roomIds")
    List<UnreadCount> findByUserIdAndRoomIds(@Param("userId") Integer userId, @Param("roomIds") List<Integer> roomIds);
}

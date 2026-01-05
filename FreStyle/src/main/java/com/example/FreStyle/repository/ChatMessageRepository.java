package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.ChatMessage;
import com.example.FreStyle.entity.ChatRoom;

@Repository
public interface ChatMessageRepository extends JpaRepository<ChatMessage, Integer> {
  List<ChatMessage> findByRoomOrderByCreatedAtAsc(ChatRoom room);
  
  /**
   * 指定ルームの最新メッセージを取得
   */
  @Query("SELECT cm FROM ChatMessage cm WHERE cm.room.id = :roomId ORDER BY cm.createdAt DESC LIMIT 1")
  Optional<ChatMessage> findLatestMessageByRoomId(@Param("roomId") Integer roomId);
  
  /**
   * 複数ルームの最新メッセージを一括取得（サブクエリで各ルームの最新メッセージIDを取得）
   */
  @Query("""
      SELECT cm FROM ChatMessage cm 
      WHERE cm.id IN (
          SELECT MAX(cm2.id) FROM ChatMessage cm2 
          WHERE cm2.room.id IN :roomIds 
          GROUP BY cm2.room.id
      )
      """)
  List<ChatMessage> findLatestMessagesByRoomIds(@Param("roomIds") List<Integer> roomIds);
}

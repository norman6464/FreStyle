package com.example.FreStyle.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.ChatMessage;
import com.example.FreStyle.entity.ChatRoom;

@Repository
public interface ChatMessageRepository extends JpaRepository<ChatMessage, Integer> {
  List<ChatMessage> findByRoomOrderByCreatedAtAsc(ChatRoom room);
}

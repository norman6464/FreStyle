package com.example.FreStyle.service;

import org.springframework.stereotype.Service;

import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.repository.ChatRoomRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class ChatRoomService {
  
  private final ChatRoomRepository chatRoomRepository;
  
  // チャットルームの取得をする
  public ChatRoom findChatRoomById(Integer roomId) {
    return chatRoomRepository.findById(roomId).orElseThrow(() -> new RuntimeException("ルームが存在しません。"));
  }
  
}

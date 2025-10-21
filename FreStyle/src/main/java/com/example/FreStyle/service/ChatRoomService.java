package com.example.FreStyle.service;

import org.springframework.stereotype.Service;

import com.example.FreStyle.repository.ChatRoomRepository;

@Service
public class ChatRoomService {
  private ChatRoomRepository chatRoomRepository;
  
  public ChatRoomService(ChatRoomRepository chatRoomRepository) {
    this.chatRoomRepository = chatRoomRepository;
  }
  
  
  
}

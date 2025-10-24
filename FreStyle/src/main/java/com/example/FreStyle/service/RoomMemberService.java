package com.example.FreStyle.service;

import java.util.List;

import org.springframework.stereotype.Service;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;

// ユーザー同士のチャットの管理
@Service
public class RoomMemberService {
  
  private final RoomMemberRepository roomMemberRepository;
  
  public RoomMemberService(RoomMemberRepository roomMemberRepository) {
    this.roomMemberRepository = roomMemberRepository;
  }
  
  public boolean existsRoom(Integer roomId, Integer userId) {
    return roomMemberRepository.existsByRoom_IdAndUser_Id(roomId, userId);
  }
  
  // ルームが一つもなかったらコントローラ側で処理を途中で検索を終了させる
  public List<Integer> findRoomId(Integer userId) {
    List<Integer> rooms = roomMemberRepository.findRoomIdByUserId(userId);
    return rooms;
  }
  
  public List<User> findUsers(Integer userId) {
    List<User> users = roomMemberRepository.findUsersByUserId(userId);
    return users;
  }
}

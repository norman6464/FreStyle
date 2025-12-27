package com.example.FreStyle.service;
import java.util.List;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.RoomMember;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.repository.UserRepository;
import lombok.RequiredArgsConstructor;

// ChatRoomServiceとRoomMemberServiceクラス二つとも関与しているときはこちらのクラスを使う
@Service
@RequiredArgsConstructor
public class ChatService {
    private final ChatRoomRepository chatRoomRepository;
    private final RoomMemberRepository roomMemberRepository;
    private final UserRepository userRepository;

    
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
}

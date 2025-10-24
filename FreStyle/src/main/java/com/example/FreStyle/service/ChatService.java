package com.example.FreStyle.service;

import java.util.List;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.RoomMember;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.repository.UserRepository;

import jakarta.annotation.PostConstruct;
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials;
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;


// ChatRoomServiceとRoomMemberServiceクラス二つとも関与しているときはこちらのクラスを使う
@Service
public class ChatService {
    @Value("${aws.access-key}")
    private String accessKey;
    
    @Value("${aws.secret-key}")
    private String secretKey;
    
    @Value("${aws.region}")
    private String region;
    
    @Value("${aws.dynamodb.table-name.chat}")
    private String tableName;
    
    private DynamoDbClient dynamoDbClient;
    
    private final ChatRoomRepository chatRoomRepository;
    private final RoomMemberRepository roomMemberRepository;
    private final UserRepository userRepository;
    
    public ChatService(ChatRoomRepository chatRoomRepository,RoomMemberRepository roomMemberRepository,UserRepository userRepository) {
      this.chatRoomRepository = chatRoomRepository;
      this.roomMemberRepository = roomMemberRepository;
      this.userRepository = userRepository;
    }
    
    
    // PostConstructではBeanが生成されて依存製注入（DI）が完了した直後に実行されるメソッドを定義する
    @PostConstruct
    public void init() {
      dynamoDbClient = DynamoDbClient.builder()
            .region(Region.of(region))
            .credentialsProvider(
              StaticCredentialsProvider.create(
                  AwsBasicCredentials.create(accessKey, secretKey)
              ) 
            )
            .build();
    }
    
    // チャットルームの作成
    @Transactional
    public Integer createOrGetRoom(Integer myUserId, Integer targetUserId) {
      // 既存ルームがあるか確認
      Integer existingRoomId = chatRoomRepository.findRoomIdByUserIds(myUserId, targetUserId);
      if (existingRoomId != null) {
        return existingRoomId;
      }
      
      // 新規チャットルームを作成
      ChatRoom newRoom = new ChatRoom();
      chatRoomRepository.save(newRoom);
      
      // 自分を追加
      RoomMember myMember = new RoomMember();
      myMember.setRoom(newRoom);
      myMember.setUser(userRepository.findById(myUserId)
              .orElseThrow(()-> new IllegalStateException("ユーザーが存在しません。")));
      
      // 相手を追加
      RoomMember targetMember = new RoomMember();
      targetMember.setRoom(newRoom);
      targetMember.setUser(userRepository.findById(targetUserId)
                  .orElseThrow(() -> new IllegalStateException("相手ユーザーが存在しません。")));
      
      roomMemberRepository.saveAll(List.of(myMember, targetMember));
                  
      return newRoom.getId();
      
    }
     
    
}

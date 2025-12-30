package com.example.FreStyle.controller;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/chat/")
@RequiredArgsConstructor
public class ChatController {

  
  private final UserService userService;
  private final ChatService chatService;
  private final ChatRoomService chatRoomService;
  private final ChatMessageService chatMessageService; 
  private final UserIdentityService userIdentityService;

  // ユーザー登録一覧
  @GetMapping("/users")
  public ResponseEntity<?> users(@AuthenticationPrincipal Jwt jwt,
      @RequestParam(name = "query", required = false) String query) {
    System.out.println("GET /api/chat/users");
    String cognitoSub = jwt.getSubject();

    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }

    User myUser = userIdentityService.findUserBySub(cognitoSub);

    List<UserDto> users = userService.findUsersWithRoomId(myUser.getId(), query);

    Map<String, List<UserDto>> responseData = new HashMap<>();

    for (UserDto user : users) {
      System.out.println("User_id" + user.getId() + "User_Email" + user.getEmail() + "User_name" + user.getName());
    }
    responseData.put("users", users);
    return ResponseEntity.ok().body(responseData);
  }

  @PostMapping("/users/{id}/create")
  public ResponseEntity<?> create(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "id") Integer id) {
    
    String cognitoSub = jwt.getSubject();
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      System.out.println("request bad request");
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なリクエストです。"));
    }
    
    try{
      
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      
      Integer roomId = chatService.createOrGetRoom(myUser.getId(), id);
      System.out.println("request ok");
      return ResponseEntity.ok(Map.of(
            "roomId", roomId,
            "status", "success"
      ));
  } catch (IllegalStateException e) {
    System.out.println(e.getMessage());
    return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
  } catch (Exception e) {
    System.out.println(e.getMessage());
    e.printStackTrace();
    return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
              .body(Map.of("error", "ルーム作成中にエラーが発生しました。"));
  } 
}

  
  @GetMapping("/users/{roomId}/history")
  public ResponseEntity<?> history(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "roomId") Integer roomId) {
    System.out.println("Request receive: roomId=" + roomId);
    
    String cognitoSub = jwt.getSubject();
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
    }
    
    try {
      // 自分のユーザー情報を取得
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      System.out.println("myUser: " + myUser.getName());
      
      // すでにroom_idが取得されている状態なのでchatRoomServiceからChatRoomオブジェクトを取得をする
      ChatRoom chatRoom = chatRoomService.findChatRoomById(roomId);
      System.out.println("chatRoom found: " + chatRoom.getId());
      
      // 履歴の取得
      List<ChatMessageDto> history = chatMessageService.getMessagesByRoom(chatRoom);
      System.out.println("history count: " + history.size());
      
      return ResponseEntity.ok(history);
      
      
    } catch (Exception e) {
      System.out.println("Error in history endpoint: " + e.getMessage());
      e.printStackTrace();
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of("error", "サーバーエラーです。"));
    }
    
  }

  @GetMapping("/stats")
  public ResponseEntity<?> stats(@AuthenticationPrincipal Jwt jwt) {
    String cognitoSub = jwt.getSubject();
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }

    try {
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      Long totalUsers = userService.getTotalUserCount();
      
      Map<String, Object> stats = new HashMap<>();
      stats.put("totalUsers", totalUsers);
      stats.put("email", myUser.getEmail());
      stats.put("username", myUser.getName());
      
      return ResponseEntity.ok().body(stats);
    } catch (Exception e) {
      System.out.println(e.getMessage());
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of("error", "サーバーエラーです。"));
    }
  }

}

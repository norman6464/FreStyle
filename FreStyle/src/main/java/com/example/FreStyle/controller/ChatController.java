package com.example.FreStyle.controller;

import java.util.ArrayList;
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
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.UserService;

@RestController
@RequestMapping("/api/chat/")
public class ChatController {

  
  private final UserService userService;
  private final ChatService chatService;

  public ChatController(UserService userService,ChatService chatService) {
    this.userService = userService;
    this.chatService = chatService;
  }

  // ユーザー登録一覧
  @GetMapping("/users")
  public ResponseEntity<?> users(@AuthenticationPrincipal Jwt jwt,
      @RequestParam(name = "query", required = false) String query) {
    System.out.println("GET /api/chat/users");
    String cognitoSub = jwt.getSubject();

    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
    }

    Integer userId = userService.findUserIdByCognitoSub(cognitoSub);

    List<UserDto> users = userService.findUsersWithRoomId(userId, query);

    Map<String, List<UserDto>> responseData = new HashMap<>();

    for (UserDto user : users) {
      System.out.println("User_id" + user.getId() + "User_Email" + user.getEmail());
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
      
      Integer userId = userService.findUserIdByCognitoSub(cognitoSub);
      
      Integer roomId = chatService.createOrGetRoom(userId, id);
      System.out.println("request ok");
      return ResponseEntity.ok(Map.of(
            "roomId", roomId,
            "status", "success"
      ));
  } catch (IllegalStateException e) {
    return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
  } catch (Exception e) {
    e.printStackTrace();
    return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
              .body(Map.of("error", "ルーム作成中にエラーが発生しました。"));
  } 
}

  
  @GetMapping("/users/{roomId}/history")
  public ResponseEntity<?> history(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "roomId") Integer roomId) {
    System.out.println("Request receive");
    
    String cognitoSub = jwt.getSubject();
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
    }
    
    try {
      
      List<ChatMessageDto> history = chatService.getChatHistory(roomId, cognitoSub);
      return ResponseEntity.ok().body(history);
      
    } catch (Exception e) {
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of("error", "サーバーエラーです。"));
    }
    
  }

}

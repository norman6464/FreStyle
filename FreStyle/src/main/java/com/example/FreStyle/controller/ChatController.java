package com.example.FreStyle.controller;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.entity.User;

import com.example.FreStyle.service.RoomMemberService;
import com.example.FreStyle.service.UserService;

// AiChatControllerと違うのは各ユーザーごとにチャットの管理をしているため
// 複雑な条件式を汲む必要がある
@RestController
@RequestMapping("/api/chat/")
public class ChatController {

  private final RoomMemberService roomMemberService;
  private final UserService userService;

  public ChatController(RoomMemberService roomMemberService, UserService userService) {
    this.roomMemberService = roomMemberService;
    this.userService = userService;
  }

  // GET api/chat/members
  // 現在のユーザーの友達一覧を表示するためのデータを返す
  @GetMapping("/members")
  public ResponseEntity<?> members(@AuthenticationPrincipal Jwt jwt) {

    // Jwtからsubを取得をする
    String cognitoSub = jwt.getSubject();

    if (cognitoSub == null || cognitoSub.trim().isEmpty()) {
      Map<String, String> errorData = new HashMap<>();
      errorData.put("error", "無効なリクエストです。");
      return ResponseEntity.badRequest().body(errorData);
    }

    try {

      List<String> usersName = new ArrayList<>();
      Integer userId = userService.findUserIdByCognitoSub(cognitoSub);
      List<Integer> roomId = roomMemberService.findRoomId(userId);
      if (!roomId.isEmpty()) {
        List<User> users = roomMemberService.findUsers(userId);
        for (User user : users) {
          usersName.add(user.getUsername());
        }
      }

      Map<String, List<String>> responseData = new HashMap<>();
      responseData.put("name", usersName);
      return ResponseEntity.ok().body(responseData);

    } catch (Exception e) {
      return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
    }

  }
}

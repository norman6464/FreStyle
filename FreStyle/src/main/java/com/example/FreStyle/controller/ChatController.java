package com.example.FreStyle.controller;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.User;

import com.example.FreStyle.service.RoomMemberService;
import com.example.FreStyle.service.UserService;

// members（友達追加をしている状態）
// users（友達追加をしていない状態）
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
  // @GetMapping("/members")
  // public ResponseEntity<?> members(@AuthenticationPrincipal Jwt jwt,
  // @RequestParam(name = "query", required = false) String query) {

  // String cognitoSub = jwt.getSubject();

  // if (cognitoSub == null || cognitoSub.trim().isEmpty()) {
  // Map<String, String> errorData = new HashMap<>();
  // errorData.put("error", "無効なリクエストです。");
  // return ResponseEntity.badRequest().body(errorData);
  // }

  // try {

  // List<String> usersName = new ArrayList<>();
  // Integer userId = userService.findUserIdByCognitoSub(cognitoSub);
  // List<Integer> roomId = roomMemberService.findRoomId(userId);
  // if (!roomId.isEmpty()) {
  // List<User> users = roomMemberService.findUsers(userId);
  // for (User user : users) {
  // usersName.add(user.getUsername());
  // }
  // }

  // Map<String, List<String>> responseData = new HashMap<>();
  // responseData.put("name", usersName);
  // return ResponseEntity.ok().body(responseData);

  // } catch (Exception e) {
  // return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
  // }

  // }

  // ユーザー登録一覧
  @GetMapping("/users")
  public ResponseEntity<?> users(@AuthenticationPrincipal Jwt jwt,
      @RequestParam(name = "query", required = false) String query) {
    String cognito_sub = jwt.getSubject();

    if (cognito_sub == null || cognito_sub.isEmpty()) {
      return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
    }

    Integer userId = userService.findUserIdByCognitoSub(cognito_sub);
    
    List<UserDto> users = userService.findUsersWithRoomId(userId, query);
    
    Map<String, List<UserDto>> responseData = new HashMap<>();

    for (UserDto user : users) {
      System.out.println("User_id" + user.getId() + "User_Email" + user.getEmail());
    }

    responseData.put("users", users);
    return ResponseEntity.ok().body(responseData);
  }

  @GetMapping("/members/{id}/")
  public ResponseEntity<?> chat(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "id") String id) {

    return ResponseEntity.ok("200 OK テストテスト");

  }

}

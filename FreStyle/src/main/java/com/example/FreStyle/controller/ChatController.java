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
import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.RoomMemberService;
import com.example.FreStyle.service.UnreadCountService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequestMapping("/api/chat/")
@RequiredArgsConstructor
@Slf4j
public class ChatController {

  
  private final UserService userService;
  private final ChatService chatService;
  private final ChatRoomService chatRoomService;
  private final ChatMessageService chatMessageService; 
  private final UserIdentityService userIdentityService;
  private final RoomMemberService roomMemberService;
  private final UnreadCountService unreadCountService;

  // ユーザー登録一覧
  @GetMapping("/users")
  public ResponseEntity<?> users(@AuthenticationPrincipal Jwt jwt,
      @RequestParam(name = "query", required = false) String query) {
    log.info("GET /api/chat/users");
    String cognitoSub = jwt.getSubject();

    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }

    User myUser = userIdentityService.findUserBySub(cognitoSub);

    List<UserDto> users = userService.findUsersWithRoomId(myUser.getId(), query);

    Map<String, List<UserDto>> responseData = new HashMap<>();

    for (UserDto user : users) {
      log.info("User_id" + user.getId() + "User_Email" + user.getEmail() + "User_name" + user.getName());
    }
    responseData.put("users", users);
    return ResponseEntity.ok().body(responseData);
  }

  @PostMapping("/users/{id}/create")
  public ResponseEntity<?> create(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "id") Integer id) {
    
    log.info("\n========== ルーム作成リクエスト開始 ==========");
    log.info("📌 リクエストPathVariable id: " + id);
    log.info("📌 JWT null判定: " + (jwt == null ? "NULL" : "存在"));
    
    String cognitoSub = jwt.getSubject();
    log.info("📌 cognitoSub (Cognito User ID): " + cognitoSub);
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      log.error("❌ cognitoSubがnullまたは空です");
      log.info("========== ルーム作成リクエスト終了(UNAUTHORIZED) ==========\n");
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なリクエストです。"));
    }
    
    try{
      log.info("🔍 userIdentityService.findUserBySub() 実行中...");
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      log.info("✅ 現在のユーザー取得成功");
      log.debug("   - myUser.getId(): " + myUser.getId());
      log.debug("   - myUser.getName(): " + myUser.getName());
      log.debug("   - myUser.getEmail(): " + myUser.getEmail());
      
      log.info("🔍 chatService.createOrGetRoom() 実行中...");
      log.debug("   - myUser.getId(): " + myUser.getId() + " (ログイン中のユーザーID)");
      log.debug("   - id (相手ユーザーID): " + id);
      Integer roomId = chatService.createOrGetRoom(myUser.getId(), id);
      
      log.info("✅ ルーム作成/取得成功");
      log.debug("   - roomId: " + roomId);
      log.info("========== ルーム作成リクエスト終了(OK) ==========\n");
      return ResponseEntity.ok(Map.of(
            "roomId", roomId,
            "status", "success"
      ));
  } catch (IllegalStateException e) {
    log.warn("ルーム作成エラー(不正リクエスト): {}", e.getMessage(), e);
    return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
  } catch (Exception e) {
    log.error("ルーム作成エラー: {}", e.getMessage(), e);
    return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
              .body(Map.of("error", "ルーム作成中にエラーが発生しました。"));
  } 
}

  
  @GetMapping("/users/{roomId}/history")
  public ResponseEntity<?> history(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "roomId", required = true) Integer roomId) {

    String cognitoSub = jwt.getSubject();
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
    }
    
    try {
      // 自分のユーザー情報を取得
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      Integer myUserId = myUser.getId();
      
      // すでにroom_idが取得されている状態なのでchatRoomServiceからChatRoomオブジェクトを取得をする
      ChatRoom chatRoom = chatRoomService.findChatRoomById(roomId);
      log.info("chatRoom found: " + chatRoom.getId());
      
      // 履歴の取得 - 現在のユーザーIDを渡す
      List<ChatMessageDto> history = chatMessageService.getMessagesByRoom(chatRoom, myUserId);
      log.info("history count: " + history.size());
      
      return ResponseEntity.ok(history);
      
    } catch (Exception e) {
      log.error("履歴取得エラー: {}", e.getMessage(), e);
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
      // 会話したことのあるユーザー数を取得
      Long chatPartnerCount = roomMemberService.countChatPartners(myUser.getId());
      
      Map<String, Object> stats = new HashMap<>();
      stats.put("chatPartnerCount", chatPartnerCount);
      stats.put("email", myUser.getEmail());
      stats.put("username", myUser.getName());
      
      return ResponseEntity.ok().body(stats);
    } catch (Exception e) {
      log.info(e.getMessage());
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of("error", "サーバーエラーです。"));
    }
  }

  /**
   * チャットルームの未読数をリセット（既読にする）
   */
  @PostMapping("/rooms/{roomId}/read")
  public ResponseEntity<?> markAsRead(
      @AuthenticationPrincipal Jwt jwt,
      @PathVariable("roomId") Integer roomId) {

    if (jwt == null) {
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "認証されていません。"));
    }

    String cognitoSub = jwt.getSubject();
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }

    try {
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      unreadCountService.resetUnreadCount(myUser.getId(), roomId);
      return ResponseEntity.ok(Map.of("status", "success"));
    } catch (Exception e) {
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
          .body(Map.of("error", "サーバーエラーが発生しました。"));
    }
  }

  /**
   * チャット履歴のあるユーザー一覧を取得
   * @param jwt 認証トークン
   * @param query 検索クエリ（名前またはメールで検索、オプション）
   * @return チャット履歴のあるユーザー一覧（最終メッセージ情報付き）
   */
  @GetMapping("/rooms")
  public ResponseEntity<?> getChatRooms(
      @AuthenticationPrincipal Jwt jwt,
      @RequestParam(name = "query", required = false) String query) {
    
    log.info("\n========== GET /api/chat/rooms ==========");
    log.info("📌 query: " + query);
    
    if (jwt == null) {
      log.error("❌ 認証エラー: JWTがnull");
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }
    
    String cognitoSub = jwt.getSubject();
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      log.error("❌ 認証エラー: cognitoSubがnullまたは空");
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }
    
    try {
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      log.info("✅ ユーザー取得成功 - ID: " + myUser.getId() + ", Name: " + myUser.getName());
      
      List<ChatUserDto> chatUsers = chatService.findChatUsers(myUser.getId(), query);
      log.info("✅ チャットユーザー取得成功 - 件数: " + chatUsers.size());
      
      Map<String, Object> response = new HashMap<>();
      response.put("chatUsers", chatUsers);
      
      log.info("========== GET /api/chat/rooms 完了 ==========\n");
      return ResponseEntity.ok(response);
      
    } catch (Exception e) {
      log.error("チャットルーム取得エラー: {}", e.getMessage(), e);
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
          .body(Map.of("error", "サーバーエラーが発生しました。"));
    }
  }

}

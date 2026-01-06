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
  private final RoomMemberService roomMemberService;

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
    
    System.out.println("\n========== ルーム作成リクエスト開始 ==========");
    System.out.println("📌 リクエストPathVariable id: " + id);
    System.out.println("📌 JWT null判定: " + (jwt == null ? "NULL" : "存在"));
    
    String cognitoSub = jwt.getSubject();
    System.out.println("📌 cognitoSub (Cognito User ID): " + cognitoSub);
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      System.out.println("❌ cognitoSubがnullまたは空です");
      System.out.println("========== ルーム作成リクエスト終了(UNAUTHORIZED) ==========\n");
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なリクエストです。"));
    }
    
    try{
      System.out.println("🔍 userIdentityService.findUserBySub() 実行中...");
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      System.out.println("✅ 現在のユーザー取得成功");
      System.out.println("   - myUser.getId(): " + myUser.getId());
      System.out.println("   - myUser.getName(): " + myUser.getName());
      System.out.println("   - myUser.getEmail(): " + myUser.getEmail());
      
      System.out.println("🔍 chatService.createOrGetRoom() 実行中...");
      System.out.println("   - myUser.getId(): " + myUser.getId() + " (ログイン中のユーザーID)");
      System.out.println("   - id (相手ユーザーID): " + id);
      Integer roomId = chatService.createOrGetRoom(myUser.getId(), id);
      
      System.out.println("✅ ルーム作成/取得成功");
      System.out.println("   - roomId: " + roomId);
      System.out.println("========== ルーム作成リクエスト終了(OK) ==========\n");
      return ResponseEntity.ok(Map.of(
            "roomId", roomId,
            "status", "success"
      ));
  } catch (IllegalStateException e) {
    System.out.println("⚠️ IllegalStateException発生: " + e.getMessage());
    System.out.println("   スタックトレース:");
    e.printStackTrace();
    System.out.println("========== ルーム作成リクエスト終了(BAD_REQUEST) ==========\n");
    return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
  } catch (Exception e) {
    System.out.println("❌ 予期しない例外発生: " + e.getClass().getName());
    System.out.println("   メッセージ: " + e.getMessage());
    System.out.println("   スタックトレース:");
    e.printStackTrace();
    System.out.println("========== ルーム作成リクエスト終了(INTERNAL_SERVER_ERROR) ==========\n");
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
      System.out.println("chatRoom found: " + chatRoom.getId());
      
      // 履歴の取得 - 現在のユーザーIDを渡す
      List<ChatMessageDto> history = chatMessageService.getMessagesByRoom(chatRoom, myUserId);
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
      // 会話したことのあるユーザー数を取得
      Long chatPartnerCount = roomMemberService.countChatPartners(myUser.getId());
      
      Map<String, Object> stats = new HashMap<>();
      stats.put("chatPartnerCount", chatPartnerCount);
      stats.put("email", myUser.getEmail());
      stats.put("username", myUser.getName());
      
      return ResponseEntity.ok().body(stats);
    } catch (Exception e) {
      System.out.println(e.getMessage());
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of("error", "サーバーエラーです。"));
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
    
    System.out.println("\n========== GET /api/chat/rooms ==========");
    System.out.println("📌 query: " + query);
    
    if (jwt == null) {
      System.out.println("❌ 認証エラー: JWTがnull");
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }
    
    String cognitoSub = jwt.getSubject();
    
    if (cognitoSub == null || cognitoSub.isEmpty()) {
      System.out.println("❌ 認証エラー: cognitoSubがnullまたは空");
      return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
          .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }
    
    try {
      User myUser = userIdentityService.findUserBySub(cognitoSub);
      System.out.println("✅ ユーザー取得成功 - ID: " + myUser.getId() + ", Name: " + myUser.getName());
      
      List<ChatUserDto> chatUsers = chatService.findChatUsers(myUser.getId(), query);
      System.out.println("✅ チャットユーザー取得成功 - 件数: " + chatUsers.size());
      
      Map<String, Object> response = new HashMap<>();
      response.put("chatUsers", chatUsers);
      
      System.out.println("========== GET /api/chat/rooms 完了 ==========\n");
      return ResponseEntity.ok(response);
      
    } catch (Exception e) {
      System.out.println("❌ エラー発生: " + e.getMessage());
      e.printStackTrace();
      return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
          .body(Map.of("error", "サーバーエラーが発生しました。"));
    }
  }

}

package com.example.FreStyle.controller;

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
import com.example.FreStyle.usecase.CreateOrGetChatRoomUseCase;
import com.example.FreStyle.usecase.GetChatHistoryUseCase;
import com.example.FreStyle.usecase.GetChatRoomsUseCase;
import com.example.FreStyle.usecase.GetChatStatsUseCase;
import com.example.FreStyle.usecase.GetChatUsersUseCase;
import com.example.FreStyle.usecase.MarkChatAsReadUseCase;

import lombok.RequiredArgsConstructor;

@RestController
@RequestMapping("/api/chat/")
@RequiredArgsConstructor
public class ChatController {

    private final GetChatUsersUseCase getChatUsersUseCase;
    private final CreateOrGetChatRoomUseCase createOrGetChatRoomUseCase;
    private final GetChatHistoryUseCase getChatHistoryUseCase;
    private final GetChatStatsUseCase getChatStatsUseCase;
    private final MarkChatAsReadUseCase markChatAsReadUseCase;
    private final GetChatRoomsUseCase getChatRoomsUseCase;

    @GetMapping("/users")
    public ResponseEntity<?> users(@AuthenticationPrincipal Jwt jwt,
        @RequestParam(name = "query", required = false) String query) {
        String cognitoSub = jwt.getSubject();
        if (cognitoSub == null || cognitoSub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
        }
        List<UserDto> users = getChatUsersUseCase.execute(cognitoSub, query);
        return ResponseEntity.ok().body(Map.of("users", users));
    }

    @PostMapping("/users/{id}/create")
    public ResponseEntity<?> create(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "id") Integer id) {
        String cognitoSub = jwt.getSubject();
        if (cognitoSub == null || cognitoSub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なリクエストです。"));
        }
        try {
            Integer roomId = createOrGetChatRoomUseCase.execute(cognitoSub, id);
            return ResponseEntity.ok(Map.of("roomId", roomId, "status", "success"));
        } catch (IllegalStateException e) {
            return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
        } catch (Exception e) {
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
            List<ChatMessageDto> history = getChatHistoryUseCase.execute(cognitoSub, roomId);
            return ResponseEntity.ok(history);
        } catch (Exception e) {
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
            Map<String, Object> stats = getChatStatsUseCase.execute(cognitoSub);
            return ResponseEntity.ok().body(stats);
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of("error", "サーバーエラーです。"));
        }
    }

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
            markChatAsReadUseCase.execute(cognitoSub, roomId);
            return ResponseEntity.ok(Map.of("status", "success"));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }

    @GetMapping("/rooms")
    public ResponseEntity<?> getChatRooms(
        @AuthenticationPrincipal Jwt jwt,
        @RequestParam(name = "query", required = false) String query) {
        if (jwt == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
        }
        String cognitoSub = jwt.getSubject();
        if (cognitoSub == null || cognitoSub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
        }
        try {
            List<ChatUserDto> chatUsers = getChatRoomsUseCase.execute(cognitoSub, query);
            return ResponseEntity.ok(Map.of("chatUsers", chatUsers));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "サーバーエラーが発生しました。"));
        }
    }
}

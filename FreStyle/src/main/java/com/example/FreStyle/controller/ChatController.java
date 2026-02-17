package com.example.FreStyle.controller;

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

import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.usecase.CreateOrGetChatRoomUseCase;
import com.example.FreStyle.usecase.GetChatHistoryUseCase;
import com.example.FreStyle.usecase.GetChatRoomsUseCase;
import com.example.FreStyle.usecase.GetChatStatsUseCase;
import com.example.FreStyle.usecase.GetChatUsersUseCase;
import com.example.FreStyle.usecase.MarkChatAsReadUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequestMapping("/api/chat/")
@RequiredArgsConstructor
@Slf4j
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
        String sub = jwt.getSubject();
        log.info("チャットユーザー一覧取得: sub={}", sub);
        List<UserDto> users = getChatUsersUseCase.execute(sub, query);
        return ResponseEntity.ok().body(Map.of("users", users));
    }

    @PostMapping("/users/{id}/create")
    public ResponseEntity<?> create(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "id") Integer id) {
        String sub = jwt.getSubject();
        log.info("チャットルーム作成: sub={}, targetId={}", sub, id);
        Integer roomId = createOrGetChatRoomUseCase.execute(sub, id);
        return ResponseEntity.ok(Map.of("roomId", roomId, "status", "success"));
    }

    @GetMapping("/users/{roomId}/history")
    public ResponseEntity<?> history(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "roomId", required = true) Integer roomId) {
        String sub = jwt.getSubject();
        log.info("チャット履歴取得: sub={}, roomId={}", sub, roomId);
        return ResponseEntity.ok(getChatHistoryUseCase.execute(sub, roomId));
    }

    @GetMapping("/stats")
    public ResponseEntity<?> stats(@AuthenticationPrincipal Jwt jwt) {
        String sub = jwt.getSubject();
        log.info("チャット統計取得: sub={}", sub);
        return ResponseEntity.ok().body(getChatStatsUseCase.execute(sub));
    }

    @PostMapping("/rooms/{roomId}/read")
    public ResponseEntity<?> markAsRead(
        @AuthenticationPrincipal Jwt jwt,
        @PathVariable("roomId") Integer roomId) {
        String sub = jwt.getSubject();
        log.info("既読マーク: sub={}, roomId={}", sub, roomId);
        markChatAsReadUseCase.execute(sub, roomId);
        return ResponseEntity.ok(Map.of("status", "success"));
    }

    @GetMapping("/rooms")
    public ResponseEntity<?> getChatRooms(
        @AuthenticationPrincipal Jwt jwt,
        @RequestParam(name = "query", required = false) String query) {
        String sub = jwt.getSubject();
        log.info("チャットルーム一覧取得: sub={}", sub);
        List<ChatUserDto> chatUsers = getChatRoomsUseCase.execute(sub, query);
        return ResponseEntity.ok(Map.of("chatUsers", chatUsers));
    }
}

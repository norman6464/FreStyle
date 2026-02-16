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
        String sub = resolveSubject(jwt);
        if (sub == null) return unauthorizedResponse();
        List<UserDto> users = getChatUsersUseCase.execute(sub, query);
        return ResponseEntity.ok().body(Map.of("users", users));
    }

    @PostMapping("/users/{id}/create")
    public ResponseEntity<?> create(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "id") Integer id) {
        String sub = resolveSubject(jwt);
        if (sub == null) return unauthorizedResponse();
        try {
            Integer roomId = createOrGetChatRoomUseCase.execute(sub, id);
            return ResponseEntity.ok(Map.of("roomId", roomId, "status", "success"));
        } catch (IllegalStateException e) {
            return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @GetMapping("/users/{roomId}/history")
    public ResponseEntity<?> history(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "roomId", required = true) Integer roomId) {
        String sub = resolveSubject(jwt);
        if (sub == null) return unauthorizedResponse();
        try {
            return ResponseEntity.ok(getChatHistoryUseCase.execute(sub, roomId));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @GetMapping("/stats")
    public ResponseEntity<?> stats(@AuthenticationPrincipal Jwt jwt) {
        String sub = resolveSubject(jwt);
        if (sub == null) return unauthorizedResponse();
        try {
            return ResponseEntity.ok().body(getChatStatsUseCase.execute(sub));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @PostMapping("/rooms/{roomId}/read")
    public ResponseEntity<?> markAsRead(
        @AuthenticationPrincipal Jwt jwt,
        @PathVariable("roomId") Integer roomId) {
        String sub = resolveSubject(jwt);
        if (sub == null) return unauthorizedResponse();
        try {
            markChatAsReadUseCase.execute(sub, roomId);
            return ResponseEntity.ok(Map.of("status", "success"));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @GetMapping("/rooms")
    public ResponseEntity<?> getChatRooms(
        @AuthenticationPrincipal Jwt jwt,
        @RequestParam(name = "query", required = false) String query) {
        String sub = resolveSubject(jwt);
        if (sub == null) return unauthorizedResponse();
        try {
            List<ChatUserDto> chatUsers = getChatRoomsUseCase.execute(sub, query);
            return ResponseEntity.ok(Map.of("chatUsers", chatUsers));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    private String resolveSubject(Jwt jwt) {
        if (jwt == null) return null;
        String sub = jwt.getSubject();
        return (sub == null || sub.isEmpty()) ? null : sub;
    }

    private ResponseEntity<Map<String, String>> unauthorizedResponse() {
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
            .body(Map.of("error", "タイムアウトをしたか、または未ログインです。"));
    }

    private ResponseEntity<Map<String, String>> internalErrorResponse() {
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
            .body(Map.of("error", "サーバーエラーが発生しました。"));
    }
}

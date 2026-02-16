package com.example.FreStyle.controller;

import java.util.List;
import java.util.Map;
import java.util.Optional;

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
        Optional<String> subject = resolveSubject(jwt);
        if (subject.isEmpty()) return unauthorizedResponse();
        List<UserDto> users = getChatUsersUseCase.execute(subject.get(), query);
        return ResponseEntity.ok().body(Map.of("users", users));
    }

    @PostMapping("/users/{id}/create")
    public ResponseEntity<?> create(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "id") Integer id) {
        Optional<String> subject = resolveSubject(jwt);
        if (subject.isEmpty()) return unauthorizedResponse();
        try {
            Integer roomId = createOrGetChatRoomUseCase.execute(subject.get(), id);
            return ResponseEntity.ok(Map.of("roomId", roomId, "status", "success"));
        } catch (IllegalStateException e) {
            return ResponseEntity.badRequest().body(Map.of("error", "無効なリクエストです。"));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @GetMapping("/users/{roomId}/history")
    public ResponseEntity<?> history(@AuthenticationPrincipal Jwt jwt, @PathVariable(name = "roomId", required = true) Integer roomId) {
        Optional<String> subject = resolveSubject(jwt);
        if (subject.isEmpty()) return unauthorizedResponse();
        try {
            return ResponseEntity.ok(getChatHistoryUseCase.execute(subject.get(), roomId));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @GetMapping("/stats")
    public ResponseEntity<?> stats(@AuthenticationPrincipal Jwt jwt) {
        Optional<String> subject = resolveSubject(jwt);
        if (subject.isEmpty()) return unauthorizedResponse();
        try {
            return ResponseEntity.ok().body(getChatStatsUseCase.execute(subject.get()));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @PostMapping("/rooms/{roomId}/read")
    public ResponseEntity<?> markAsRead(
        @AuthenticationPrincipal Jwt jwt,
        @PathVariable("roomId") Integer roomId) {
        Optional<String> subject = resolveSubject(jwt);
        if (subject.isEmpty()) return unauthorizedResponse();
        try {
            markChatAsReadUseCase.execute(subject.get(), roomId);
            return ResponseEntity.ok(Map.of("status", "success"));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    @GetMapping("/rooms")
    public ResponseEntity<?> getChatRooms(
        @AuthenticationPrincipal Jwt jwt,
        @RequestParam(name = "query", required = false) String query) {
        Optional<String> subject = resolveSubject(jwt);
        if (subject.isEmpty()) return unauthorizedResponse();
        try {
            List<ChatUserDto> chatUsers = getChatRoomsUseCase.execute(subject.get(), query);
            return ResponseEntity.ok(Map.of("chatUsers", chatUsers));
        } catch (Exception e) {
            return internalErrorResponse();
        }
    }

    private Optional<String> resolveSubject(Jwt jwt) {
        if (jwt == null) return Optional.empty();
        String sub = jwt.getSubject();
        return (sub == null || sub.isEmpty()) ? Optional.empty() : Optional.of(sub);
    }

    private ResponseEntity<Map<String, String>> unauthorizedResponse() {
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
            .body(Map.of("error", "タイムアウトしたか、または未ログインです。"));
    }

    private ResponseEntity<Map<String, String>> internalErrorResponse() {
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
            .body(Map.of("error", "サーバーエラーが発生しました。"));
    }
}

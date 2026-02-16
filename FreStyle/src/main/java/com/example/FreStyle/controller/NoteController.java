package com.example.FreStyle.controller;

import com.example.FreStyle.dto.NoteDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CreateNoteUseCase;
import com.example.FreStyle.usecase.DeleteNoteUseCase;
import com.example.FreStyle.usecase.GetNotesByUserIdUseCase;
import com.example.FreStyle.usecase.UpdateNoteUseCase;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/notes")
@Slf4j
public class NoteController {
    private final GetNotesByUserIdUseCase getNotesByUserIdUseCase;
    private final CreateNoteUseCase createNoteUseCase;
    private final UpdateNoteUseCase updateNoteUseCase;
    private final DeleteNoteUseCase deleteNoteUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<List<NoteDto>> getNotes(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);

        List<NoteDto> notes = getNotesByUserIdUseCase.execute(user.getId());
        log.info("ノート一覧取得成功 - 件数: {}", notes.size());

        return ResponseEntity.ok(notes);
    }

    @PostMapping
    public ResponseEntity<NoteDto> createNote(
            @AuthenticationPrincipal Jwt jwt,
            @RequestBody CreateNoteRequest request
    ) {
        User user = resolveUser(jwt);

        NoteDto note = createNoteUseCase.execute(user.getId(), request.title());
        log.info("ノート作成成功 - noteId: {}", note.getNoteId());

        return ResponseEntity.ok(note);
    }

    @PutMapping("/{noteId}")
    public ResponseEntity<Void> updateNote(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable String noteId,
            @RequestBody UpdateNoteRequest request
    ) {
        User user = resolveUser(jwt);

        updateNoteUseCase.execute(user.getId(), noteId, request.title(), request.content(), request.isPinned());
        log.info("ノート更新成功 - noteId: {}", noteId);

        return ResponseEntity.noContent().build();
    }

    @DeleteMapping("/{noteId}")
    public ResponseEntity<Void> deleteNote(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable String noteId
    ) {
        User user = resolveUser(jwt);

        deleteNoteUseCase.execute(user.getId(), noteId);
        log.info("ノート削除成功 - noteId: {}", noteId);

        return ResponseEntity.noContent().build();
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }

    public record CreateNoteRequest(String title) {}
    public record UpdateNoteRequest(String title, String content, Boolean isPinned) {}
}

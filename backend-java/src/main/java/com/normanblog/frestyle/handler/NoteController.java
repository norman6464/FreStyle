package com.normanblog.frestyle.handler;

import com.normanblog.frestyle.domain.Note;
import com.normanblog.frestyle.usecase.NoteService;
import jakarta.validation.constraints.NotBlank;
import java.util.List;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * NoteController はノート API の HTTP 入口(handler 層)。
 * 認証はまだ未移植のため、暫定で userId=1 固定。Cognito JWT middleware 移植時に差し替える。
 * Go 版 handler.NoteHandler と互換(GET/POST /api/v2/notes)。
 */
@RestController
@RequestMapping("/api/v2/notes")
public class NoteController {

  private static final Long TEMP_USER_ID = 1L;

  private final NoteService noteService;

  public NoteController(NoteService noteService) {
    this.noteService = noteService;
  }

  @GetMapping
  public List<Note> list() {
    return noteService.listByUser(TEMP_USER_ID);
  }

  @PostMapping
  public Note create(@RequestBody CreateNoteRequest req) {
    return noteService.create(TEMP_USER_ID, req.title(), req.content());
  }

  /** CreateNoteRequest はノート作成リクエストボディ。 */
  public record CreateNoteRequest(@NotBlank String title, String content) {}
}

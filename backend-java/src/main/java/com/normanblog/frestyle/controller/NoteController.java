package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.CreateNoteRequest;
import com.normanblog.frestyle.dto.NoteResponse;
import com.normanblog.frestyle.service.NoteService;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** ノート API のエンドポイントを公開するコントローラ。 */
@RestController
@RequestMapping("/api/v2/notes")
public class NoteController {

  // 認証(Cognito JWT)が未実装の段階でも notes の動作とメモリ実測を進められるよう、
  // 暫定で固定ユーザーにしている。認証導入時に認証情報から userId を取得する形へ差し替える。
  private static final Long TEMP_USER_ID = 1L;

  private final NoteService noteService;

  public NoteController(NoteService noteService) {
    this.noteService = noteService;
  }

  @GetMapping
  public List<NoteResponse> list() {
    return noteService.findByUser(TEMP_USER_ID).stream().map(NoteResponse::from).toList();
  }

  @PostMapping
  public NoteResponse create(@Valid @RequestBody CreateNoteRequest request) {
    return NoteResponse.from(
        noteService.create(TEMP_USER_ID, request.title(), request.content()));
  }
}

package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.CreateNoteRequest;
import com.normanblog.frestyle.dto.NoteResponse;
import com.normanblog.frestyle.security.CurrentUserProvider;
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

  private final NoteService noteService;
  private final CurrentUserProvider currentUser;

  public NoteController(NoteService noteService, CurrentUserProvider currentUser) {
    this.noteService = noteService;
    this.currentUser = currentUser;
  }

  @GetMapping
  public List<NoteResponse> list() {
    Long userId = currentUser.require().getId();
    return noteService.findByUser(userId).stream().map(NoteResponse::from).toList();
  }

  @PostMapping
  public NoteResponse create(@Valid @RequestBody CreateNoteRequest request) {
    Long userId = currentUser.require().getId();
    return NoteResponse.from(noteService.create(userId, request.title(), request.content()));
  }
}

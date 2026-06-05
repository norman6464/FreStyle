package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.Note;
import com.normanblog.frestyle.repository.NoteRepository;
import java.time.Instant;
import java.util.List;
import org.springframework.stereotype.Service;

/** ノートに関するビジネスロジックを担うサービス。 */
@Service
public class NoteService {

  private final NoteRepository noteRepository;

  public NoteService(NoteRepository noteRepository) {
    this.noteRepository = noteRepository;
  }

  public List<Note> findByUser(Long userId) {
    return noteRepository.findByUserIdOrderByIsPinnedDescUpdatedAtDesc(userId);
  }

  public Note create(Long userId, String title, String content) {
    Instant now = Instant.now();
    Note note =
        Note.builder()
            .userId(userId)
            .title(title)
            // content は任意項目。null のまま保存すると一覧取得側で null チェックが各所に
            // 散らばるため、ここで空文字に正規化して下流の分岐を不要にする。
            .content(content == null ? "" : content)
            .isPublic(false)
            .isPinned(false)
            .createdAt(now)
            .updatedAt(now)
            .build();
    return noteRepository.save(note);
  }
}

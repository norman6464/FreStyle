package com.normanblog.frestyle.usecase;

import com.normanblog.frestyle.domain.Note;
import com.normanblog.frestyle.repository.NoteRepository;
import java.time.Instant;
import java.util.List;
import org.springframework.stereotype.Service;

/**
 * NoteService はノートのビジネスロジック(Application 層)。
 * handler から呼ばれ、repository をオーケストレーションする。
 * FreStyle(Go) の usecase 群に対応(1 ユースケース=1 メソッドで段階的に分割予定)。
 */
@Service
public class NoteService {

  private final NoteRepository notes;

  public NoteService(NoteRepository notes) {
    this.notes = notes;
  }

  public List<Note> listByUser(Long userId) {
    return notes.findByUserIdOrderByIsPinnedDescUpdatedAtDesc(userId);
  }

  public Note create(Long userId, String title, String content) {
    Instant now = Instant.now();
    Note note =
        Note.builder()
            .userId(userId)
            .title(title == null || title.isBlank() ? "無題" : title)
            .content(content == null ? "" : content)
            .isPublic(false)
            .isPinned(false)
            .createdAt(now)
            .updatedAt(now)
            .build();
    return notes.save(note);
  }
}

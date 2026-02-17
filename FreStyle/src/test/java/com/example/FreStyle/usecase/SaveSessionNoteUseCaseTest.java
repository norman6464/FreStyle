package com.example.FreStyle.usecase;

import static org.mockito.ArgumentMatchers.argThat;
import static org.mockito.Mockito.*;

import java.time.LocalDateTime;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.SessionNote;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.SessionNoteRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("SaveSessionNoteUseCase")
class SaveSessionNoteUseCaseTest {

    @Mock
    private SessionNoteRepository sessionNoteRepository;

    @InjectMocks
    private SaveSessionNoteUseCase useCase;

    @Test
    @DisplayName("新規メモを作成する")
    void createsNewNote() {
        User user = new User();
        user.setId(1);
        when(sessionNoteRepository.findByUserIdAndSessionId(1, 100))
                .thenReturn(Optional.empty());

        useCase.execute(user, 100, "新規メモ");

        verify(sessionNoteRepository).save(argThat(n ->
                n.getSessionId() == 100 && n.getNote().equals("新規メモ") && n.getUser().getId() == 1));
    }

    @Test
    @DisplayName("既存メモを更新する")
    void updatesExistingNote() {
        User user = new User();
        user.setId(1);
        SessionNote existing = new SessionNote();
        existing.setId(10);
        existing.setUser(user);
        existing.setSessionId(100);
        existing.setNote("旧メモ");
        existing.setUpdatedAt(LocalDateTime.of(2026, 2, 15, 10, 0));
        when(sessionNoteRepository.findByUserIdAndSessionId(1, 100))
                .thenReturn(Optional.of(existing));

        useCase.execute(user, 100, "更新メモ");

        verify(sessionNoteRepository).save(argThat(n ->
                n.getId() == 10 && n.getNote().equals("更新メモ")));
    }

    @Test
    @DisplayName("既存メモ更新時にUserとSessionIdが変更されない")
    void updatesExistingNotePreservesUserAndSessionId() {
        User user = new User();
        user.setId(1);
        SessionNote existing = new SessionNote();
        existing.setId(10);
        existing.setUser(user);
        existing.setSessionId(100);
        existing.setNote("旧メモ");
        when(sessionNoteRepository.findByUserIdAndSessionId(1, 100))
                .thenReturn(Optional.of(existing));

        useCase.execute(user, 100, "更新メモ");

        verify(sessionNoteRepository).save(argThat(n ->
                n.getUser().getId() == 1 && n.getSessionId() == 100));
    }

    @Test
    @DisplayName("異なるセッションIDで別のメモが新規作成される")
    void createsNewNoteWithDifferentSessionId() {
        User user = new User();
        user.setId(1);
        when(sessionNoteRepository.findByUserIdAndSessionId(1, 200))
                .thenReturn(Optional.empty());

        useCase.execute(user, 200, "別セッションメモ");

        verify(sessionNoteRepository).save(argThat(n ->
                n.getSessionId() == 200 && n.getNote().equals("別セッションメモ")));
    }
}

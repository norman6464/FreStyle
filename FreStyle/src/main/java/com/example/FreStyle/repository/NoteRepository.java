package com.example.FreStyle.repository;

import com.example.FreStyle.dto.NoteDto;

import java.util.List;

public interface NoteRepository {

    List<NoteDto> findByUserId(Integer userId);

    NoteDto save(Integer userId, String title);

    void update(Integer userId, String noteId, String title, String content, Boolean isPinned);

    void delete(Integer userId, String noteId);
}

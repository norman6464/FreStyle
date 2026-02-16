package com.example.FreStyle.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.FavoritePhrase;

@Repository
public interface FavoritePhraseRepository extends JpaRepository<FavoritePhrase, Integer> {
    List<FavoritePhrase> findByUserIdOrderByCreatedAtDesc(Integer userId);

    boolean existsByUserIdAndRephrasedTextAndPattern(Integer userId, String rephrasedText, String pattern);

    void deleteByIdAndUserId(Integer id, Integer userId);
}

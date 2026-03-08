package com.example.FreStyle.repository;

import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import com.example.FreStyle.entity.ConversationTemplate;

@Repository
public interface ConversationTemplateRepository extends JpaRepository<ConversationTemplate, Integer> {
    List<ConversationTemplate> findByCategory(String category);
    List<ConversationTemplate> findByDifficulty(String difficulty);
}

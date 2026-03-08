package com.example.FreStyle.usecase;

import java.util.List;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.example.FreStyle.dto.ConversationTemplateDto;
import com.example.FreStyle.entity.ConversationTemplate;
import com.example.FreStyle.repository.ConversationTemplateRepository;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class GetConversationTemplatesUseCase {
    private final ConversationTemplateRepository repository;

    public List<ConversationTemplateDto> execute(String category) {
        List<ConversationTemplate> templates = (category == null || category.isEmpty())
            ? repository.findAll()
            : repository.findByCategory(category);

        return templates.stream()
            .map(t -> new ConversationTemplateDto(
                t.getId(), t.getTitle(), t.getDescription(),
                t.getCategory(), t.getOpeningMessage(), t.getDifficulty()))
            .toList();
    }
}

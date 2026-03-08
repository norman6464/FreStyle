package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import com.example.FreStyle.entity.ConversationTemplate;
import com.example.FreStyle.repository.ConversationTemplateRepository;
import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class GetConversationTemplateByIdUseCase {
    private final ConversationTemplateRepository repository;

    public ConversationTemplate execute(Integer id) {
        return repository.findById(id)
            .orElseThrow(() -> new IllegalArgumentException("テンプレートが見つかりません: " + id));
    }
}

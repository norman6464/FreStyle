package com.example.FreStyle.controller;

import java.util.List;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.*;
import com.example.FreStyle.dto.ConversationTemplateDto;
import com.example.FreStyle.usecase.GetConversationTemplatesUseCase;
import com.example.FreStyle.usecase.GetConversationTemplateByIdUseCase;
import com.example.FreStyle.entity.ConversationTemplate;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/templates")
@Slf4j
public class ConversationTemplateController {
    private final GetConversationTemplatesUseCase getTemplatesUseCase;
    private final GetConversationTemplateByIdUseCase getTemplateByIdUseCase;

    @GetMapping
    public ResponseEntity<List<ConversationTemplateDto>> getTemplates(
            @AuthenticationPrincipal Jwt jwt,
            @RequestParam(required = false) String category
    ) {
        log.info("========== GET /api/templates?category={} ==========", category);
        List<ConversationTemplateDto> templates = getTemplatesUseCase.execute(category);
        log.info("テンプレート一覧取得成功 - 件数: {}", templates.size());
        return ResponseEntity.ok(templates);
    }

    @GetMapping("/{id}")
    public ResponseEntity<ConversationTemplateDto> getTemplateById(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer id
    ) {
        log.info("========== GET /api/templates/{} ==========", id);
        ConversationTemplate t = getTemplateByIdUseCase.execute(id);
        ConversationTemplateDto dto = new ConversationTemplateDto(
            t.getId(), t.getTitle(), t.getDescription(),
            t.getCategory(), t.getOpeningMessage(), t.getDifficulty());
        return ResponseEntity.ok(dto);
    }
}

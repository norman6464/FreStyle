package com.example.FreStyle.mapper;

import java.util.Objects;

import org.springframework.stereotype.Component;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;

/**
 * 練習シナリオのマッピングクラス
 *
 * <p>役割:</p>
 * <ul>
 *   <li>PracticeScenarioエンティティ ⇔ PracticeScenarioDTOの相互変換</li>
 *   <li>プレゼンテーション層とドメイン層の境界を明確化</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>プレゼンテーション層とアプリケーション層の間のマッピング層</li>
 *   <li>DTOとEntityの変換ロジックを一箇所に集約</li>
 * </ul>
 */
@Component
public class PracticeScenarioMapper {

    /**
     * エンティティからDTOへ変換
     *
     * @param entity 練習シナリオエンティティ
     * @return 練習シナリオDTO（APIレスポンス用）
     * @throws NullPointerException entityがnullの場合
     */
    public PracticeScenarioDto toDto(PracticeScenario entity) {
        Objects.requireNonNull(entity, "PracticeScenarioエンティティがnullです");

        return new PracticeScenarioDto(
            entity.getId(),
            entity.getName(),
            entity.getDescription(),
            entity.getCategory(),
            entity.getRoleName(),
            entity.getDifficulty(),
            entity.getSystemPrompt()
        );
    }

    /**
     * DTOからエンティティへ変換
     *
     * <p>注意:</p>
     * <ul>
     *   <li>IDは新規作成時はnullとなる（DBの自動採番に依存）</li>
     *   <li>systemPromptはDTOに含まれないため、別途設定が必要</li>
 *   <li>createdAtはDBのデフォルト値に依存</li>
     * </ul>
     *
     * @param dto 練習シナリオDTO
     * @return 練習シナリオエンティティ
     * @throws NullPointerException dtoがnullの場合
     */
    public PracticeScenario toEntity(PracticeScenarioDto dto) {
        Objects.requireNonNull(dto, "PracticeScenarioDTOがnullです");

        PracticeScenario entity = new PracticeScenario();
        entity.setId(dto.id());
        entity.setName(dto.name());
        entity.setDescription(dto.description());
        entity.setCategory(dto.category());
        entity.setRoleName(dto.roleName());
        entity.setDifficulty(dto.difficulty());
        entity.setSystemPrompt(dto.systemPrompt());

        return entity;
    }
}

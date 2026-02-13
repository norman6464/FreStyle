package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.mapper.PracticeScenarioMapper;
import com.example.FreStyle.repository.PracticeScenarioRepository;

import lombok.RequiredArgsConstructor;

/**
 * ID指定練習シナリオ取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>指定されたIDの練習シナリオをデータベースから取得</li>
 *   <li>エンティティをDTOに変換してプレゼンテーション層に返却</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>指定されたIDが存在しない場合はRuntimeExceptionをスロー</li>
 *   <li>エラーメッセージには日本語で分かりやすく記載</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 *   <li>単一のエンティティ取得という単純な責務</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class GetPracticeScenarioByIdUseCase {

    private final PracticeScenarioRepository practiceScenarioRepository;
    private final PracticeScenarioMapper mapper;

    /**
     * ID指定で練習シナリオを取得
     *
     * @param id 練習シナリオID
     * @return 練習シナリオDTO
     * @throws ResourceNotFoundException 指定されたIDのシナリオが見つからない場合
     */
    @Transactional(readOnly = true)
    public PracticeScenarioDto execute(Integer id) {
        PracticeScenario entity = practiceScenarioRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("シナリオが見つかりません: ID=" + id));

        return mapper.toDto(entity);
    }
}

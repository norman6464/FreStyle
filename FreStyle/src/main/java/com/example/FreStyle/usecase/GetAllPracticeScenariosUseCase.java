package com.example.FreStyle.usecase;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.PracticeScenarioDto;
import com.example.FreStyle.mapper.PracticeScenarioMapper;
import com.example.FreStyle.repository.PracticeScenarioRepository;

import lombok.RequiredArgsConstructor;

/**
 * 全練習シナリオ取得ユースケース
 *
 * <p>役割:</p>
 * <ul>
 *   <li>すべての練習シナリオをデータベースから取得</li>
 *   <li>エンティティをDTOに変換してプレゼンテーション層に返却</li>
 * </ul>
 *
 * <p>ビジネスルール:</p>
 * <ul>
 *   <li>全シナリオを一括取得（フィルタリングなし）</li>
 *   <li>カテゴリ別フィルタリングは別のUseCaseで実装</li>
 * </ul>
 *
 * <p>クリーンアーキテクチャー上の位置づけ:</p>
 * <ul>
 *   <li>アプリケーション層（Use Case層）</li>
 *   <li>ビジネスロジックを含まない単純なCRUD操作</li>
 * </ul>
 */
@Service
@RequiredArgsConstructor
public class GetAllPracticeScenariosUseCase {

    private final PracticeScenarioRepository practiceScenarioRepository;
    private final PracticeScenarioMapper mapper;

    /**
     * 全練習シナリオを取得
     *
     * @return 練習シナリオDTOのリスト
     */
    @Transactional(readOnly = true)
    public List<PracticeScenarioDto> execute() {
        return practiceScenarioRepository.findAll().stream()
                .map(mapper::toDto)
                .collect(Collectors.toList());
    }
}

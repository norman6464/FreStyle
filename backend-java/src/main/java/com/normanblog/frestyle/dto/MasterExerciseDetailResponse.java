package com.normanblog.frestyle.dto;

import java.util.List;

/** GET /exercises/{slug} のレスポンス。演習本体 + 例の配列。 */
public record MasterExerciseDetailResponse(
    MasterExerciseResponse exercise, List<MasterExerciseExampleResponse> examples) {}

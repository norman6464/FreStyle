package com.example.FreStyle.dto;

import java.time.LocalDateTime;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * エラーレスポンスDTO
 *
 * <p>役割:</p>
 * <ul>
 *   <li>APIエラー時の統一されたレスポンス形式を提供</li>
 *   <li>クライアントがエラーを適切に処理できるよう必要な情報を含む</li>
 * </ul>
 *
 * <p>フィールド:</p>
 * <ul>
 *   <li>timestamp: エラー発生日時</li>
 *   <li>status: HTTPステータスコード</li>
 *   <li>error: HTTPステータスの説明</li>
 *   <li>message: エラーメッセージ（ユーザーに表示可能）</li>
 *   <li>path: エラーが発生したリクエストパス</li>
 * </ul>
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ErrorResponseDto {
    private LocalDateTime timestamp;
    private int status;
    private String error;
    private String message;
    private String path;

    /**
     * エラーレスポンスを作成（ファクトリーメソッド）
     *
     * @param status HTTPステータスコード
     * @param error HTTPステータスの説明
     * @param message エラーメッセージ
     * @param path リクエストパス
     * @return ErrorResponseDto
     */
    public static ErrorResponseDto of(int status, String error, String message, String path) {
        return new ErrorResponseDto(
            LocalDateTime.now(),
            status,
            error,
            message,
            path
        );
    }
}

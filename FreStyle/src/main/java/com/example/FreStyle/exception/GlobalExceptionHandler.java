package com.example.FreStyle.exception;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.context.request.WebRequest;

import com.example.FreStyle.dto.ErrorResponseDto;

/**
 * グローバル例外ハンドラー
 *
 * <p>役割:</p>
 * <ul>
 *   <li>アプリケーション全体の例外を一箇所で処理</li>
 *   <li>統一されたエラーレスポンス形式を提供</li>
 *   <li>各Controllerの例外処理コードを削減</li>
 * </ul>
 *
 * <p>ミドルウェアとしての機能:</p>
 * <ul>
 *   <li>@RestControllerAdviceにより全Controllerに適用</li>
 *   <li>例外発生時に自動的にこのハンドラーが呼ばれる</li>
 *   <li>適切なHTTPステータスコードとエラーメッセージを返却</li>
 * </ul>
 */
@RestControllerAdvice
public class GlobalExceptionHandler {

    private static final Logger logger = LoggerFactory.getLogger(GlobalExceptionHandler.class);

    /**
     * ResourceNotFoundException のハンドラー
     *
     * <p>リソースが見つからない場合に404 Not Foundを返却</p>
     *
     * @param ex ResourceNotFoundException
     * @param request WebRequest
     * @return エラーレスポンス（404 Not Found）
     */
    @ExceptionHandler(ResourceNotFoundException.class)
    public ResponseEntity<ErrorResponseDto> handleResourceNotFoundException(
            ResourceNotFoundException ex,
            WebRequest request
    ) {
        logger.warn("❌ リソースが見つかりません: {}", ex.getMessage());

        ErrorResponseDto error = ErrorResponseDto.of(
            HttpStatus.NOT_FOUND.value(),
            "Not Found",
            ex.getMessage(),
            request.getDescription(false).replace("uri=", "")
        );

        return ResponseEntity.status(HttpStatus.NOT_FOUND).body(error);
    }

    /**
     * UnauthorizedException のハンドラー
     *
     * <p>権限エラーの場合に403 Forbiddenを返却</p>
     *
     * @param ex UnauthorizedException
     * @param request WebRequest
     * @return エラーレスポンス（403 Forbidden）
     */
    @ExceptionHandler(UnauthorizedException.class)
    public ResponseEntity<ErrorResponseDto> handleUnauthorizedException(
            UnauthorizedException ex,
            WebRequest request
    ) {
        logger.warn("❌ 権限エラー: {}", ex.getMessage());

        ErrorResponseDto error = ErrorResponseDto.of(
            HttpStatus.FORBIDDEN.value(),
            "Forbidden",
            ex.getMessage(),
            request.getDescription(false).replace("uri=", "")
        );

        return ResponseEntity.status(HttpStatus.FORBIDDEN).body(error);
    }

    /**
     * BusinessException のハンドラー
     *
     * <p>ビジネスロジックエラーの場合に400 Bad Requestを返却</p>
     *
     * @param ex BusinessException
     * @param request WebRequest
     * @return エラーレスポンス（400 Bad Request）
     */
    @ExceptionHandler(BusinessException.class)
    public ResponseEntity<ErrorResponseDto> handleBusinessException(
            BusinessException ex,
            WebRequest request
    ) {
        logger.warn("❌ ビジネスロジックエラー: {}", ex.getMessage());

        ErrorResponseDto error = ErrorResponseDto.of(
            HttpStatus.BAD_REQUEST.value(),
            "Bad Request",
            ex.getMessage(),
            request.getDescription(false).replace("uri=", "")
        );

        return ResponseEntity.status(HttpStatus.BAD_REQUEST).body(error);
    }

    /**
     * 汎用Exception のハンドラー
     *
     * <p>予期しないエラーの場合に500 Internal Server Errorを返却</p>
     *
     * @param ex Exception
     * @param request WebRequest
     * @return エラーレスポンス（500 Internal Server Error）
     */
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponseDto> handleGenericException(
            Exception ex,
            WebRequest request
    ) {
        logger.error("❌ 予期しないエラー: {}", ex.getMessage(), ex);

        ErrorResponseDto error = ErrorResponseDto.of(
            HttpStatus.INTERNAL_SERVER_ERROR.value(),
            "Internal Server Error",
            "サーバーエラーが発生しました。管理者にお問い合わせください。",
            request.getDescription(false).replace("uri=", "")
        );

        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(error);
    }
}

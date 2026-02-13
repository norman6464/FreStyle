package com.example.FreStyle.exception;

/**
 * ビジネスロジックエラーの場合にスローされる例外
 *
 * <p>用途:</p>
 * <ul>
 *   <li>ビジネスルール違反が発生した場合</li>
 *   <li>入力値が業務要件を満たさない場合</li>
 *   <li>システム的には正常だが、業務的にエラーとなる場合</li>
 * </ul>
 *
 * <p>HTTPステータスコード: 400 Bad Request</p>
 */
public class BusinessException extends RuntimeException {

    /**
     * メッセージ付きでBusinessExceptionを作成
     *
     * @param message エラーメッセージ
     */
    public BusinessException(String message) {
        super(message);
    }

    /**
     * メッセージと原因例外付きでBusinessExceptionを作成
     *
     * @param message エラーメッセージ
     * @param cause 原因例外
     */
    public BusinessException(String message, Throwable cause) {
        super(message, cause);
    }
}

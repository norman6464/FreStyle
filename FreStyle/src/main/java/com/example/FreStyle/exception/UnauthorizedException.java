package com.example.FreStyle.exception;

/**
 * 権限エラーの場合にスローされる例外
 *
 * <p>用途:</p>
 * <ul>
 *   <li>ユーザーが他のユーザーのリソースにアクセスしようとした場合</li>
 *   <li>認証が必要な操作を未認証ユーザーが実行しようとした場合</li>
 * </ul>
 *
 * <p>HTTPステータスコード: 403 Forbidden</p>
 */
public class UnauthorizedException extends RuntimeException {

    /**
     * メッセージ付きでUnauthorizedExceptionを作成
     *
     * @param message エラーメッセージ
     */
    public UnauthorizedException(String message) {
        super(message);
    }

    /**
     * メッセージと原因例外付きでUnauthorizedExceptionを作成
     *
     * @param message エラーメッセージ
     * @param cause 原因例外
     */
    public UnauthorizedException(String message, Throwable cause) {
        super(message, cause);
    }
}

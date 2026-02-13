package com.example.FreStyle.exception;

/**
 * リソースが見つからない場合にスローされる例外
 *
 * <p>用途:</p>
 * <ul>
 *   <li>データベースに指定されたIDのレコードが存在しない場合</li>
 *   <li>ユーザーがアクセス権限を持たないリソースにアクセスした場合</li>
 * </ul>
 *
 * <p>HTTPステータスコード: 404 Not Found</p>
 */
public class ResourceNotFoundException extends RuntimeException {

    /**
     * メッセージ付きでResourceNotFoundExceptionを作成
     *
     * @param message エラーメッセージ
     */
    public ResourceNotFoundException(String message) {
        super(message);
    }

    /**
     * メッセージと原因例外付きでResourceNotFoundExceptionを作成
     *
     * @param message エラーメッセージ
     * @param cause 原因例外
     */
    public ResourceNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }
}

## フロントエンドセットアップ手順

1. リポジトリをクローンして `frontend` ディレクトリに移動します。

2. `npm install` を実行し、`package.json` に記載されている依存パッケージをインストールして `node_modules` を作成します。

3. `npm run dev` で動作確認を行います。

4. Tailwind CSSの動作も確認してください。  
   ※ 使用しているNode.jsのバージョンによっては、Tailwind CSSが正常に動作しない可能性があります。

5. 動作しない場合は、一度 Tailwind CSSをアンインストールします。  
   ```bash
   npm uninstall tailwindcss
   
6. そのあとにもう一度インストールをします。`npm install -D tailwindcss@バージョン指定`

7.　`npx tailwindcss init -p` で初期設定をします。

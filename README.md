git cloneの手順（frontendディレクトリ）
1.npm install でpackage.jsonに記載されているのでnode_modulesディレクトリを作成する

2.npm dev run で動作確認

3.tailwindcssの動作も確認（使用しているnode.jsのバージョンとの安定板のtailwindcssの動作しない可能性がある為）

4.もし動作しない場合はnpm uninstall tailwindcss

5.npm install -D tailwindcss@指定バージョン

6.npx tailwindcss  init -p

# ISUCON12 Template

ISUCON12用リポジトリテンプレート

`/home/isucon`に展開する前提

```bash
task -g --list # Taskfile.yamlにある主要コマンド一覧
task -g {コマンド} # コマンドの実行
```

`-g`をつけると、どこのディレクトリで実行しても`~/Taskfile.yaml`が参照される

## taskコマンドを実行できるまでにセットアップ

```bash!
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
```

## taskコマンド編集時のメモ

`task --list`に出てくるのは`desc`があるものだけ

なので、本番中に叩く前提のコマンドには`desc`を書き、間接的に呼ばれるだけのものは`desc`を書かないようにしておく
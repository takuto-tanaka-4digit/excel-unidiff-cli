# excel-unidiff-cli (`euni`)

[English](./README.md) | [日本語](./README.ja.md)

Unicode 正規化（NFC/NFD）起因の Git 幽霊差分を、検知と修復で抑止するクロスプラットフォーム CLI。macOS/Linux/Windows で同じ挙動を再現でき、CI でも安定運用できます。

## 問題

macOS/Linux/Windows 混在チームでは、Unicode 正規化差で次が起きます。

- discard 後も `git status` が dirty のまま
- 実質同じファイル名なのに差分が出続ける
- ローカルと CI で判定が食い違う

特に、Excel を多用するリポジトリ、多言語ファイル名、日本語/記号/空白を含む環境で顕在化しやすいです。

## なぜ重要か（チーム運用・CI）

共通ポリシーがないと、Unicode drift は運用コストに直結します。

- PR ノイズ増加: 意味のないファイル名差分が混入
- CI 不安定化: Runner 依存で fail/pass が揺れる
- 自動化の停滞: Bot やスクリプトが dirty 状態で再試行を繰り返す
- 調査時間増大: 根本原因の切り分けが難しい

`euni` は、単一ポリシーと明確な契約で、Runner をまたいだ再現性を確保します。

## `euni` でできること

- Unicode 安全性に関わる Git config drift を検知
- drift を `git config --local` のみで修復
- NFC/NFD リスク、結合文字、衝突を解析
- findings と実行エラーを明確分離（終了コード安定）
- CI 向け機械可読 JSON レポートを出力

## コマンド

- `euni check`
  - 非破壊の検証
  - drift + Unicode 健全性を検査（追跡済みファイルのみ）
- `euni apply`
  - drift 修復（`git config --local`）
  - `--dry-run` 対応
  - `--repair-unicode-deletes` 対応（`core.precomposeunicode=false` モードで、ステージ済み/作業ツリーの削除済み追跡ファイルを復旧）
- `euni doctor`
  - `check` + `scan` を統合して原因分類
- `euni scan`
  - 追跡済み + 未追跡ファイルの Unicode パス解析
- `euni init-policy`
  - `.euni.yml` テンプレート生成
- `euni version`
  - バージョンとコミット表示

## 対象範囲と影響マトリクス

| Command | `--recursive=false`（既定） | `--recursive=true` | 書き込み |
| --- | --- | --- | --- |
| `check` | root repo のみ | root + submodule + nested repo | なし |
| `apply` | root repo のみ | root + submodule + nested repo | あり（`git config --local`; `--repair-unicode-deletes` 指定時は index/worktree も更新） |
| `doctor` | root repo のみ | root + submodule + nested repo | なし |
| `scan` | root repo のみ | root + submodule + nested repo | なし |
| `init-policy` | root repo のみ | N/A | あり（`.euni.yml`） |
| `version` | N/A | N/A | なし |

注記: subtree 配下は root のファイル走査対象として常に含まれます。

## 安全性保証と非目標

`euni` は安全側デフォルトで設計しています。

- 既定スコープは狭い（`--recursive=false`）
- 書き込みは `apply` と `init-policy` のみ
- `apply` 既定の書き込み範囲は `git config --local` のみ
- `apply --repair-unicode-deletes` 指定時は削除済み追跡パスに対して index/worktree を更新:
  - staged 削除: `git restore --staged --worktree`
  - worktree のみ削除: `git restore --worktree`（staged の非削除変更は保持）
- `apply --dry-run` で変更予定を事前確認
- global/system Git config は変更しない
- 破壊的 Git 操作は行わない
  - `git reset --hard` なし
  - `git clean -fd` なし
- v1 ではファイル名自動 rename なし

## クイックスタート

1. policy テンプレートを生成

```bash
euni init-policy --repo .
```

2. 現在状態を検査

```bash
euni check --repo . --recursive --policy ./.euni.yml
```

3. 変更予定を確認してから適用

```bash
euni apply --repo . --recursive --policy ./.euni.yml --dry-run
euni apply --repo . --recursive --policy ./.euni.yml
```

`git status` に Unicode 由来の幽霊削除差分が出る場合は、次の1コマンドで復旧:

```bash
euni apply --repo . --recursive --policy ./.euni.yml --repair-unicode-deletes
```

注意: `--repair-unicode-deletes` は対象範囲の「削除済み追跡パス」を一括復旧します。  
対象範囲に意図した削除がある場合は使わないでください。

4. 切り分けが難しいケースを診断

```bash
euni doctor --repo . --recursive --policy ./.euni.yml
```

## Homebrew でインストール（同一リポジトリ Tap）

このリポジトリ自体を Tap として利用します。

```bash
brew tap --custom-remote takuto-tanaka-4digit/excel-unidiff-cli https://github.com/takuto-tanaka-4digit/excel-unidiff-cli
brew install takuto-tanaka-4digit/excel-unidiff-cli/euni
```

更新時:

```bash
brew update
brew upgrade takuto-tanaka-4digit/excel-unidiff-cli/euni
```

## CI 組み込みの要点

JSON モードを使い、レポート契約を検証してください。

```bash
euni check --repo . --recursive --policy ./.euni.yml --non-interactive --format json > euni-report.json
```

推奨 CI ゲート:

- `schema/euni-report.schema.json` で `euni-report.json` を検証
- プロセス終了コードと `status/exitCode` の整合検証
- findings（`1`）と実行エラー（`2`）でジョブ失敗
- `euni-report.json` を常に artifact 保存

JSON 契約:

- `stdout`: 単一 JSON オブジェクトのみ
- `stderr`: 進捗・警告・エラー詳細

## 終了コード契約

- `0`: 問題なし
- `1`: findings あり（対応要）
- `2`: 実行エラー（repo/policy/git/runtime）

実行エラーが1件でもある場合、`2` が優先されます。

## Policy ファイル

`euni` は `.euni.yml` を、root/submodule/nested repo を横断する期待値の単一情報源として扱います。

## 想定ユーザー

- macOS 開発 + Linux/Windows CI の混在チーム
- 日本語/記号/空白を含むファイル名を扱うリポジトリ
- Unicode 健全性を決定的かつ監査可能に運用したいメンテナー

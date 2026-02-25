---
name: spec
description: Excel UniDiff CLI（euni）の正式仕様。要件確認、実装判断、テスト設計、CI運用方針の参照時に使う。Unicode正規化差分（NFC/NFD）とGit設定drift対策を含む。
---

# Excel UniDiff CLI（euni）仕様書

- 文書ID: `UGC-SPEC-001`
- バージョン: `v1.0`
- 作成日: `2026-02-25`
- 対象: macOS / Linux / Windows の Git リポジトリ（submodule 含む）

## 1. 背景と目的

macOS ではファイルシステムと Git の Unicode 正規化差（NFC/NFD）により、以下が起こる。

- `discard` しても差分が消えない
- 実質差分なしなのに `git status` が dirty
- 自動化（Cursor/Codex/CI）が差分処理で停滞

本仕様の目的は、**Runner 非依存で Unicode 差分を統一管理**し、開発・CI どちらでも再現可能に抑止すること。

## 2. 命名（共通化）

- CLI リポジトリ名: `excel-unidiff-cli`
- 実行バイナリ名: `euni`
- 互換エイリアス: `excel-unidiff`（symlink/alias 推奨）
- Go module: `github.com/<org>/excel-unidiff-cli`

`ms-docs` 固有名は使わない。他リポジトリへ横展開前提。

## 3. スコープ

### 3.1 対象

- Git root リポジトリ
- 全 submodule（`--recursive` 指定時）
- root 配下の全サブフォルダ（再帰）
- Git subtree 配下のファイル群（root配下として再帰対象）
- root 配下で検出される nested Git repository（`.git` を持つディレクトリ）
- Unicode 設定 drift 検知・修復
- NFC 衝突・結合文字（combining mark）検査
### 3.2 非対象（v1）

- ファイル名自動 rename（破壊リスク高）
- `git reset --hard` / `git clean -fd` の自動実行
- Git 以外の VCS

## 4. 正常系ポリシー（基準）

Git 公式説明準拠:

- `core.precomposeUnicode=true` は macOS で分解済み文字列を再合成し、他OS共有に有効
- `core.protectHFS=true` は macOS で有効化推奨

標準基準:

- root（macOS）:
  - `core.precomposeunicode=true`
  - `core.protecthfs=true`
- root（Linux/Windows/WSL）:
  - `core.precomposeunicode=false`
- submodule:
  - デフォルトは root 準拠
  - 例外は policy で明示（例: 特定 submodule を `false` 固定）

補足:

- CI が Linux/Windows で動く場合、**CI 側でも同じ policy で check を強制**しないと再発する。
- 「mac だけ設定」で十分ではない。

## 5. CLI 要件

### 5.1 コマンド一覧

1. `euni check`
2. `euni apply`
3. `euni doctor`
4. `euni scan`
5. `euni init-policy`
6. `euni version`

### 5.2 共通オプション

- `--repo <path>`: 対象 root（必須、デフォルト `.`）
- `--recursive`: サブフォルダ以降を再帰探索（submodule / subtree / nested repo を含む）
- `--policy <path>`: policy YAML パス
- `--format text|json`: 出力形式（デフォルト `text`）
- `--quiet`: 最小出力
- `--non-interactive`: 非対話実行（プロンプト禁止）

### 5.3 サブコマンド仕様

#### check

- 目的: drift 検知（`--fix` 指定で自動修正）
- 挙動:
  - root + submodule + nested Git repository の expected/actual 比較
  - root 配下サブフォルダを再帰走査し、subtree 配下を含めて Unicode 健全性を検査
  - drift 一覧出力
  - `--fix` 指定時は drift 項目を `git config --local` で即時修正し、再評価
- 終了コード:
  - `0`: drift なし、または `--fix` で全修正完了
  - `1`: drift あり（未修正/修正失敗を含む）
  - `2`: 実行エラー

#### apply

- 目的: drift 修復
- 挙動:
  - drift 項目のみ `git config --local` で修正（root + submodule + nested Git repository）
  - 修正後に再評価
  - `--dry-run` 時は変更予定のみ
- 終了コード:
  - `0`: 全修復完了
  - `1`: 一部未修復
  - `2`: 実行エラー

#### doctor

- 目的: 失敗時の原因診断
- 挙動:
  - `check` + `scan` 相当を統合
  - 原因分類:
    - config drift
    - NFC 衝突
    - combining mark 混在
    - Git 管理外要因
  - 次アクションを1行提案

#### scan

- 目的: Unicode パス健全性解析
- 挙動:
  - root 配下サブフォルダを再帰走査
  - `git ls-files -z`（追跡済み）を対象に解析
  - `git ls-files --others --exclude-standard -z`（未追跡）も対象に解析
  - subtree 配下ファイルは root の追跡/未追跡ファイルとして解析対象に含める
  - 指標:
    - NFC-only 件数
    - NFD-only 件数
    - NFC衝突件数（`normalize(NFC)` 後に同名化）
    - combining mark 含有パス

### 5.4 再帰探索仕様（`--recursive`）

探索順:

1. `--repo` で指定された root
2. `.gitmodules` で定義された submodule（再帰）
3. root 配下サブフォルダ内で `.git` を持つ nested Git repository（再帰）
4. subtree 配下は root のファイル走査対象として自動包含

備考:

- subtree は独立 repo ではないため、config drift 判定は root 設定に準拠。
- nested Git repository は独立 repo として個別に drift 判定・修復する。

#### init-policy

- 目的: policy テンプレート生成
- 出力: `.euni.yml`

## 6. Policy ファイル仕様

ファイル名: `.euni.yml`

```yaml
version: 1
defaults:
  darwin:
    core.precomposeunicode: true
    core.protecthfs: true
  others:
    core.precomposeunicode: false
submodules:
  confluence-sync:
    core.precomposeunicode: false
nestedRepos:
  tools/legacy-repo:
    core.precomposeunicode: false
```

要件:

- キー未指定は defaults を継承
- submodule 指定は path 名で解決
- nested repository 例外指定は `nestedRepos` で path 名指定可能
- 不正キーは warning（`--non-interactive` 時は exit 2）

## 7. アーキテクチャ（Go）

### 7.1 技術

- Go `1.22+`
- Unicode 正規化: `golang.org/x/text/unicode/norm`
- CLI: `cobra` もしくは標準 `flag`（実装方針に合わせる）

### 7.2 パッケージ構成（例）

- `cmd/euni/`
- `internal/gitx/`（git 実行抽象化）
- `internal/policy/`（YAML 読込）
- `internal/drift/`（期待値/実測比較）
- `internal/scan/`（NFC/NFD 解析）
- `internal/report/`（text/json 出力）

### 7.3 Git 実行ルール

- 全コマンドは `git -C <repo>` 経由
- 対話禁止 (`GIT_TERMINAL_PROMPT=0`)
- 変更系は `--local` のみ（global/system 変更禁止）

## 8. VS Code 起動時自動実行仕様

`.vscode/tasks.json` で `runOn: folderOpen` の task に統合。

推奨コマンド:

```bash
euni check --fix --repo "${workspaceFolder}" --recursive --policy "${workspaceFolder}/.euni.yml" --non-interactive --quiet || exit 1
```

運用:

- 初回は VS Code の自動タスク許可が必要
- 既存 `safe-update` と直列化（先に `euni apply`）

## 9. CI 要件（Runner 非依存）

全 Runner で同一チェック:

```bash
euni check --repo . --recursive --policy ./.euni.yml --non-interactive --format json
```

必須要件:

- Linux / Windows / macOS すべてで実行
- drift 検知時はジョブ失敗
- JSON 結果を artifact 保存

## 10. エラー設計

- 例:
  - `UG001`: repo 不正
  - `UG002`: git 実行失敗
  - `UG003`: policy 読込失敗
  - `UG004`: drift 残存
  - `UG005`: NFC 衝突検知
- メッセージ方針:
  - ユーザー向けは平易文
  - 詳細は `--format json` で machine-readable

## 11. ログと監査

- 既定ログ: 標準出力
- `--log-file <path>` 指定で永続化
- JSON には最低限以下を含める:
  - timestamp
  - repo path
  - command
  - expected / actual
  - actions applied
  - exit code

## 12. テスト要件

### 12.1 単体

- policy パース
- drift 判定
- NFC/NFD 解析
- exit code 判定

### 12.2 結合

- root のみ
- submodule 1段/多段
- subtree を含む root 配下サブフォルダ再帰
- nested Git repository 1段/多段
- macOS/Linux/Windows それぞれで `check/apply`

### 12.3 回帰

- 「discard しても消えない差分」シナリオ
- Excel ファイル名（日本語・記号・空白）を含むケース

## 13. 受け入れ基準（DoD）

1. `apply` 後、`check` が全 Runner で成功
2. root + submodule + nested Git repository で policy と実測が一致
3. VS Code `folderOpen` 実行後に Unicode 起因の幽霊差分が再発しない
4. JSON 出力が CI で解析可能
5. 破壊的操作なしで復旧可能

## 14. 導入手順（運用）

1. `excel-unidiff-cli` リポジトリを新規作成
2. CLI をビルド・配布（Release + checksum）
3. 各対象リポジトリへ `.euni.yml` 配置
4. `tasks.json` の `folderOpen` に `euni apply` 追加
5. CI に `euni check` 追加（drift で exit 1）
6. Cursor/Codex Skill から `doctor/apply` を呼ぶ

---

この仕様は「共通CLIとして横展開する」ことを前提にしている。  
実装時は本仕様の `DoD` を満たすことを完了条件とする。

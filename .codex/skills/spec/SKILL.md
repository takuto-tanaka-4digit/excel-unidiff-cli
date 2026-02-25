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
- root repo 内ファイル走査（常に再帰）
- Git subtree 配下のファイル群（対象 repo 配下として解析対象）
- root 配下で検出される nested Git repository（`git rev-parse --show-toplevel` で判定、`--recursive=true` 時）
- Unicode 設定 drift 検知・修復
- NFC 衝突・結合文字（combining mark）検査
### 3.2 非対象（v1）

- ファイル名自動 rename（破壊リスク高）
- `git reset --hard` / `git clean -fd` の自動実行
- `git worktree` / `--separate-git-dir` で gitdir が `--repo` 境界外の構成
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

- `--repo <path>`: 対象 root（省略時 `.`）
- `--recursive`: 既定 `false`。`true` のとき submodule / nested repo を追加探索（subtree は root 走査で常時対象）
- `--policy <path>`: policy YAML パス（`check|apply|doctor` のみ。省略時 `<repo>/.euni.yml`）
- `--format text|json`: 出力形式（`check|apply|doctor|scan` のみ。デフォルト `text`）
- `--quiet`: 進捗ログ抑制（`check|apply|doctor|scan` のみ。結果レポートは維持）
- `--non-interactive`: 非対話実行（プロンプト禁止。現行コマンドでは将来互換のため受理し no-op）
- `--log-file <path>`: 監査ログ保存先（`check|apply|doctor|scan` のみ）

適用表:

| option | check/apply/doctor/scan | init-policy | version |
| --- | --- | --- | --- |
| `--repo` | 可 | 可（生成先root） | 不可 |
| `--recursive` | 可 | 不可 | 不可 |
| `--policy` | `check|apply|doctor` のみ可 | 不可 | 不可 |
| `--format` | 可 | 不可 | 不可 |
| `--quiet` | 可 | 不可 | 不可 |
| `--non-interactive` | 可 | 可 | 不可 |
| `--log-file` | 可 | 不可 | 不可 |

規約:

- 非対応オプションを渡した場合は `UG009` として `exit 2`

### 5.3 対象範囲マトリクス

| command | `--recursive=false`（既定） | `--recursive=true` |
| --- | --- | --- |
| `check` | root のみ | root + submodule + nested repo |
| `apply` | root のみ | root + submodule + nested repo |
| `doctor` | root のみ | root + submodule + nested repo |
| `scan` | root のみ | root + submodule + nested repo |
| `init-policy` | 対象外 | 対象外 |
| `version` | 対象外 | 対象外 |

補足:

- subtree は独立 repo ではない。対象 repo 配下ファイルとして自動包含。
- `scan` は各対象 repo ごとに追跡/未追跡ファイルを解析し、最終結果を集約出力する。
- root repo 内のファイル走査は常に再帰。`--recursive` は追加 repo（submodule/nested）探索の有効化だけを制御する。

### 5.4 サブコマンド仕様

#### check

- 目的: drift + Unicode 健全性検知（非破壊）
- 挙動:
  - 5.3 の対象範囲に対して expected/actual を比較
  - 対象 repo 配下を走査し、subtree 配下を含めて Unicode 健全性を検査
  - Unicode 検査対象は `git ls-files -z`（追跡済み）のみ（未追跡は `scan`/`doctor` で扱う）
  - drift 一覧出力
  - 設定変更は行わない
  - `--fix` オプションは提供しない（修復は `apply` のみ）
  - config 比較は実効 bool 値で実施（未設定は `false` とみなす）
- 終了コード:
  - `0`: findings なし
  - `1`: findings あり（例: drift / Unicode / `UG012` / `UG013`）
  - `2`: 実行エラー

#### apply

- 目的: drift 修復
- コマンド専用オプション:
  - `--dry-run`: 変更予定のみ表示（書き込みしない）
- 挙動:
  - 5.3 の対象範囲に対して drift 項目のみ `git config --local` で修正
  - 修正後に `check` 相当で再評価（drift + Unicode）
  - `--dry-run` 時は変更予定のみ表示
- 終了コード:
  - `0`: 修正後に findings なし
  - `1`: `--dry-run` で変更予定あり、または findings 残存（例: drift / Unicode / `UG012` / `UG013`）
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
- 終了コード:
  - `0`: 問題なし
  - `1`: 対応が必要な問題あり
  - `2`: 実行エラー

#### scan

- 目的: Unicode パス健全性解析
- 挙動:
  - policy は参照しない（`--policy` 非対応）
  - 5.3 の対象範囲ごとに解析
  - `git ls-files -z`（追跡済み）を対象に解析
  - `git ls-files --others --exclude-standard -z`（未追跡）も対象に解析
  - subtree 配下ファイルは対象 repo の追跡/未追跡ファイルとして解析対象に含める
  - 指標:
    - NFC-only 件数
    - NFD-only 件数
    - NFC衝突件数（`normalize(NFC)` 後に同名化）
    - combining mark 含有パス
- 終了コード:
  - `0`: Unicode 健全性問題なし
  - `1`: findings あり（NFC 衝突 / combining mark / `UG012`）
  - `2`: 実行エラー

#### init-policy

- 目的: policy テンプレート生成
- コマンド専用オプション:
  - `--force`: 既存 `.euni.yml` を上書き
- 出力: `.euni.yml`
- 挙動:
  - 出力先は `--repo` 直下
  - 既存 `.euni.yml` がある場合、`--force` なしでは上書きしない
  - `--force` 時も `.euni.yml` は no-follow で検査し、symlink/junction/reparse point の場合は `UG010` で失敗
  - `--force` 時は `.euni.yml` の実体 canonical path が `--repo` 境界内であることを必須化（境界外は `UG010`）
  - 生成 YAML はキー順を安定化して出力
- 終了コード:
  - `0`: 生成成功
  - `2`: 実行エラー（既存ファイルありで `--force` なしは `UG008`）

#### version

- 目的: バージョン表示
- 挙動:
  - `euni <semver> (<commit>)` 形式で標準出力
  - オプション受理は 5.2 適用表に従う
- 終了コード:
  - `0`: 正常終了
  - `2`: 実行エラー

### 5.5 再帰探索仕様（`--recursive`）

既定:

- `--recursive=false`。探索対象は `--repo` で指定した root のみ。

`--recursive=true` 時の探索順:

1. `--repo` で指定された root
2. `.gitmodules` と `git submodule status --recursive` で取得した submodule（再帰）
3. root 配下サブフォルダを走査し、`git -C <dir> rev-parse --show-toplevel` が成功し、かつ top-level が root と異なる nested Git repository
4. subtree 配下は root のファイル走査対象として自動包含

備考:

- subtree は独立 repo ではないため、config drift 判定は root 設定に準拠。
- nested Git repository は独立 repo として個別に drift 判定・修復する。
- submodule が未初期化/アクセス不能の場合は `UG006` として `exit 2`。
- repo 判定は canonical absolute path で重複排除し、同一 repo は1回だけ処理。
- repo 重複排除キーは Unicode 正規化を行わない（NFC/NFD を潰さない）lossless な canonical 実体パスを使う。
- canonical 化（Windows）:
  - 区切り文字を `\\` 正規化後に内部比較は `/` へ統一
  - drive letter は大文字化
  - `\\\\?\\` / UNC / 8.3 short name を同一 canonical path へ正規化
  - Unicode 文字列は保持し、repo 同一性判定では NFC/NFD 変換しない
- canonical 化（macOS/Linux）:
  - `realpath` で実体パスへ正規化
  - 区切り文字は `/` を使用
  - Unicode 文字列は保持し、repo 同一性判定では NFC/NFD 変換しない
- NFC 正規化比較は policy path 解決（6章）と JSON ソート（5.7）でのみ適用する。
- 境界比較ルール（Windows）:
  - 大文字小文字は無視して比較
  - 文字列前方一致ではなく path segment 境界で比較（`/repo` と `/repo-x` を区別）
- 安全境界チェック:
  - `--show-toplevel` と `--absolute-git-dir` の両方を canonical 化
  - 両方が `--repo` 境界内であることを必須とする
  - 境界外参照は `UG010` として `exit 2`（`git worktree` / `--separate-git-dir` 含む）
- `UG001` は `--repo` が Git repo ではない/参照不可のときに限定する。
- nested 候補は `.git`（file/dir）を持つディレクトリに限定する。
- `.git` を持つ候補で `rev-parse` 失敗時は message 文字列で分類せず `UG002` として `exit 2`（locale 非依存）。
- シンボリックリンク / junction / reparse point / mount point は再帰対象から除外し `UG012` findings を記録する。
- root 走査時は `.git/`, `.git/modules/`, `.git/worktrees/` を探索対象から除外する。

### 5.6 終了コード共通表

- `0`: 正常終了（問題なし、または期待どおりの修復完了）
- `1`: 実務上の検知結果あり（findings。drift/Unicode/管理外要因/policy 適用不能/`--dry-run` 変更予定）
- `2`: 実行エラー（repo 不正、git 実行失敗、policy 不正、submodule 未初期化/アクセス不能、内部エラー）
- 優先順位: 実行エラーが1件でもある場合は `2` を優先（findings が同時にあっても `2`）

### 5.7 JSON 契約（`--format json`）

出力チャネル:

- stdout: 単一 JSON オブジェクトのみ（機械可読）
- stderr: 進捗/警告/エラー詳細
- `--log-file` は stderr と同内容を永続化（stdout 契約は維持）

固定フィールド（必須）:

- `schemaVersion` (`string`)
- `command` (`string`)
- `repo` (`string`)
- `recursive` (`boolean`)
- `status` (`string`, `ok|findings|error`)
- `exitCode` (`integer`, `0|1|2`)
- `summary` (`object`)
- `results` (`array`)
- `errors` (`array`)
- `startedAt` (`string`, RFC3339)
- `finishedAt` (`string`, RFC3339)

`results[]` 最低限フィールド:

- `repoPath` (`string`)
- `kind` (`string`, `drift|unicode|policy|environment`)
- `code` (`string`)
- `message` (`string`)
- `path` (`string|null`)
- `targetType` (`string`, `path|configKey|repo`)
- `expected` (`string|null`)
- `actual` (`string|null`)
- `action` (`string|null`)
- `details` (`object|null`)

順序・互換:

- `results` は `repoPath` → `kind` → `path` → `code` → `expected` → `actual` の昇順で安定ソート
- `errors` は `code` → `repoPath` → `path` の昇順で安定ソート
- ソート比較は `repoPath` と `path` を canonical 化（5.5）+ NFC 正規化した UTF-8 byte order で実施
- `null` は常に最後に並べる（nulls-last）
- 互換は `schemaVersion` で管理し、破壊的変更は major を上げる
- `schemaVersion` の現行値は `"1.0.0"`（semver 形式）に固定
- `status` と `exitCode` は固定対応:
  - `ok` -> `0`
  - `findings` -> `1`
  - `error` -> `2`
- 状態不変条件:
  - `status=ok` のとき `summary.findings=0` かつ `summary.errors=0`
  - `status=findings` のとき `summary.findings>0` かつ `summary.errors=0`
  - `status=error` のとき `summary.errors>0` かつ `exitCode=2`
- `summary` 必須フィールド:
  - `targetRepos` (`integer`, 実処理対象 repo 数。findings/error が 0 件でも 1 以上になり得る)
  - `findings` (`integer`)
  - `errors` (`integer`)
  - `durationMs` (`integer`)
  - `nfcOnly` (`integer`)
  - `nfdOnly` (`integer`)
  - `nfcCollisions` (`integer`)
  - `combiningMarkPaths` (`integer`)
- `summary` 不変条件:
  - `summary.findings == results.length`
  - `summary.errors == errors.length`
  - `summary.targetRepos >= unique(nonNull(results.repoPath ∪ errors.repoPath)).length`
- `errors[]` 必須フィールド:
  - `code` (`string`)
  - `message` (`string`)
  - `repoPath` (`string|null`)
  - `path` (`string|null`)
  - `hint` (`string|null`)
- 分類ルール:
  - `results[]`: findings（`exitCode=1` 相当、例: `UG004`, `UG005`, `UG011`, `UG012`, `UG013`）
  - `errors[]`: 実行エラー（`exitCode=2` 相当、例: `UG001`, `UG002`, `UG003`, `UG006`, `UG007`, `UG008`, `UG009`, `UG010`）
- NFC 衝突（`UG005`）は `details` に次を必須で持つ:
  - `normalizedPath` (`string`)
  - `collidingPaths` (`string[]`, 2件以上)
- Schema 検証:
  - `schema/euni-report.schema.json`（JSON Schema Draft 2020-12）を正本とする
  - schema ルートの `$schema` を `https://json-schema.org/draft/2020-12/schema` に固定
  - top-level と `results[]` / `errors[]` は `additionalProperties: false` で固定
  - CI では schema 検証を必須とする（例: `ajv validate --spec=draft2020 --strict=true -s schema/euni-report.schema.json -d euni-report.json`)
  - CI では「プロセス終了コード」と `report.exitCode` / `report.status` の整合を必ず検証する
  - CI では `summary` 不変条件の整合を必ず検証する

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
- path 解決は比較前に次を適用:
  - repo 相対 path へ正規化
  - 区切り文字を `/` に統一
  - Unicode NFC 正規化
  - 大文字小文字は区別して比較（Runner 非依存で deterministic）
  - case 差のみの複数候補が存在する場合は `UG013` findings として扱い、自動解決しない
- 不正キーは常に `UG007` として `exit 2`（strict）

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
- 変更系は `rev-parse --show-toplevel` と `--absolute-git-dir` の両方が `--repo` 境界内である対象に限定

## 8. VS Code 起動時自動実行仕様

`.vscode/tasks.json` で `runOn: folderOpen` の task に統合。

推奨コマンド:

```bash
euni apply --repo "${workspaceFolder}" --recursive --policy "${workspaceFolder}/.euni.yml" --non-interactive --quiet
```

運用:

- 初回は VS Code の自動タスク許可が必要
- 既存 `safe-update` と直列化（先に `euni apply`）

## 9. CI 要件（Runner 非依存）

全 Runner で同一チェック:

1. `check` を実行し、終了コードを `check_rc` として保持（step は継続）
2. `euni-report.json` を artifact 保存（always）
3. `schema/euni-report.schema.json` で `euni-report.json` を検証し `schema_rc` を保持（always）
4. `check_rc` と `report.exitCode/report.status` の整合、および `summary` 不変条件の整合を検証し `consistency_rc` を保持（always）
5. 最終終了コードは次の優先順位で決定:
   - `schema_rc != 0` または `consistency_rc != 0` のとき `2`
   - それ以外は `check_rc`

実装例（bash）:

```bash
set +e
euni check --repo . --recursive --policy ./.euni.yml --non-interactive --format json > euni-report.json
check_rc=$?
node -e 'process.exit(/^v22\./.test(process.version)?0:2)'
node_ver_rc=$?
npx --yes ajv-cli@5.0.0 validate --spec=draft2020 --strict=true -s schema/euni-report.schema.json -d euni-report.json
schema_rc=$?
node -e 'const r=require("./euni-report.json"); const c=Number(process.argv[1]); const m={ok:0,findings:1,error:2}; const s=new Set([...(r.results||[]).map(x=>x.repoPath).filter(Boolean),...(r.errors||[]).map(x=>x.repoPath).filter(Boolean)]); const okState=(r.status==="ok"?(r.summary.findings===0&&r.summary.errors===0):(r.status==="findings"?(r.summary.findings>0&&r.summary.errors===0):(r.status==="error"?(r.summary.errors>0&&r.exitCode===2):false))); const ok=(r.schemaVersion==="1.0.0")&&(r.exitCode===c)&&(m[r.status]===r.exitCode)&&okState&&(r.summary.findings===r.results.length)&&(r.summary.errors===r.errors.length)&&(r.summary.targetRepos>=s.size); process.exit(ok?0:2)' "$check_rc"
consistency_rc=$?
set -e
# artifact upload step: always
if [ "$node_ver_rc" -ne 0 ] || [ "$schema_rc" -ne 0 ] || [ "$consistency_rc" -ne 0 ]; then exit 2; fi
exit "$check_rc"
```

実装例（PowerShell）:

```powershell
& euni check --repo . --recursive --policy ./.euni.yml --non-interactive --format json | Out-File -FilePath euni-report.json -Encoding utf8NoBOM
$check_rc = $LASTEXITCODE
$node_ver_rc = if ($PSVersionTable.PSVersion.Major -ge 7 -and (node --version) -match '^v22\.') { 0 } else { 2 }
npx --yes ajv-cli@5.0.0 validate --spec=draft2020 --strict=true -s schema/euni-report.schema.json -d euni-report.json
$schema_rc = $LASTEXITCODE
$report = Get-Content euni-report.json -Raw | ConvertFrom-Json
$statusMap = @{ ok = 0; findings = 1; error = 2 }
$repoSet = [System.Collections.Generic.HashSet[string]]::new()
foreach ($i in $report.results) { if ($null -ne $i.repoPath -and $i.repoPath -ne '') { $null = $repoSet.Add([string]$i.repoPath) } }
foreach ($i in $report.errors) { if ($null -ne $i.repoPath -and $i.repoPath -ne '') { $null = $repoSet.Add([string]$i.repoPath) } }
$okState = if ($report.status -eq 'ok') { ($report.summary.findings -eq 0) -and ($report.summary.errors -eq 0) } elseif ($report.status -eq 'findings') { ($report.summary.findings -gt 0) -and ($report.summary.errors -eq 0) } elseif ($report.status -eq 'error') { ($report.summary.errors -gt 0) -and ($report.exitCode -eq 2) } else { $false }
$ok = ($report.schemaVersion -eq '1.0.0') -and ($report.exitCode -eq $check_rc) -and ($statusMap[$report.status] -eq $report.exitCode) -and $okState -and ($report.summary.findings -eq $report.results.Count) -and ($report.summary.errors -eq $report.errors.Count) -and ($report.summary.targetRepos -ge $repoSet.Count)
$consistency_rc = if ($ok) { 0 } else { 2 }
# artifact upload step: always
if (($node_ver_rc -ne 0) -or ($schema_rc -ne 0) -or ($consistency_rc -ne 0)) { exit 2 }
exit $check_rc
```

必須要件:

- Linux / Windows / macOS すべてで実行
- findings 検知時はジョブ失敗
- JSON 結果（`euni-report.json`）を artifact 保存
- JSON 形式は 5.7 の契約に準拠し、stderr のログ混在を許可しない
- `euni-report.json` は `schema/euni-report.schema.json` で必ず検証
- `check` のプロセス終了コードと JSON `exitCode/status` の一致を必ず検証
- CI の検証ツール版を固定（例: Node 22.x, ajv-cli 5.x）

## 10. エラー設計

- 例:
  - `UG001`: repo 不正
  - `UG002`: git 実行失敗
  - `UG003`: policy 読込失敗
  - `UG004`: drift 検知（findings）
  - `UG005`: NFC 衝突検知（findings）
  - `UG006`: submodule 未初期化/アクセス不能
  - `UG007`: policy 不正キー
  - `UG008`: init-policy 上書き拒否（既存ファイルあり、`--force` なし）
  - `UG009`: 非対応オプション指定
  - `UG010`: 非対応 gitdir 構成（worktree / separate-git-dir）
  - `UG011`: combining mark 含有パス検知（findings）
  - `UG012`: Git 管理外要因検知（findings）
  - `UG013`: policy 適用不能/未解決マッチ検知（findings）
- メッセージ方針:
  - ユーザー向けは平易文
  - 詳細は `--format json` で machine-readable
  - 各 UG コードは定型 `hint` を持つ（自動化と運用Runbook参照用）

## 11. ログと監査

- `--format text`: レポートを stdout、進捗/警告/エラー詳細を stderr
- `--format json`: stdout は単一 JSON のみ、進捗/警告/エラー詳細は stderr
- `--log-file <path>`: stderr ログを永続化（stdout 契約は変更しない）
- `--quiet`: 進捗ログのみ抑制。警告/エラーは stderr 出力を維持
- text 形式のエラー行は `[UGxxx] <message> (hint: <text>)` 形式を必須とする
- JSON フィールド契約は 5.7 に準拠

## 12. テスト要件

### 12.1 単体

- policy パース
- drift 判定
- NFC/NFD 解析
- exit code 判定
- JSON Schema 検証（`schema/euni-report.schema.json`）

### 12.2 結合

- root のみ
- submodule 1段/多段
- subtree を含む root 配下サブフォルダ再帰
- nested Git repository 1段/多段
- macOS/Linux/Windows それぞれで `check/apply/doctor/scan`
- `core.ignorecase=true/false` 両方で policy path 解決

### 12.3 回帰

- 「discard しても消えない差分」シナリオ
- Excel ファイル名（日本語・記号・空白）を含むケース
- `check` は追跡済みのみ、`scan` は追跡済み+未追跡を解析する契約差分

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
5. CI に `euni check` 追加（findings で exit 1）
6. Cursor/Codex Skill から `doctor/apply` を呼ぶ

---

この仕様は「共通CLIとして横展開する」ことを前提にしている。  
実装時は本仕様の `DoD` を満たすことを完了条件とする。

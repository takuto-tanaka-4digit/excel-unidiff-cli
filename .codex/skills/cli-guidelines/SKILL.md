---
name: cli-guidelines
description: CLI 設計ガイドライン（cli-guidelines 原典の要約/参照用）。
---

# コマンドライン・インターフェース・ガイドライン

伝統的な UNIX 原則を現代向けに更新し、より良いコマンドライン・プログラムを書くための [open-source](https://github.com/cli-guidelines/cli-guidelines) ガイド。

## 著者 {#authors}

**Aanand Prasad** \
Squarespace のエンジニア、Docker Compose 共同作者。 \
[@aanandprasad](https://twitter.com/aanandprasad)

**Ben Firshman** \
[Replicate](https://replicate.com/) 共同創業者、Docker Compose 共同作者。 \
[@bfirsh](https://twitter.com/bfirsh)

**Carl Tashian** \
[Smallstep](https://smallstep.com/) の Offroad Engineer、Zipcar 初期エンジニア、Trove 共同創業者。 \
[tashian.com](https://tashian.com/) [@tashian](https://twitter.com/tashian)

**Eva Parish** \
Squarespace のテクニカルライター、O'Reilly 寄稿者。 \
[evaparish.com](https://evaparish.com/) [@evpari](https://twitter.com/evpari)

デザインは [Mark Hurrell](https://mhurrell.co.uk/)。初期貢献の Andreas Jansson、ドラフトレビューの Andrew Reitz、Ashley Williams、Brendan Falk、Chester Ramey、Dj Walker-Morgan、Jacob Maine、James Coglan、Michael Dwan、Steve Klabnik に感謝。

<iframe class="github-button" src="https://ghbtns.com/github-btn.html?user=cli-guidelines&repo=cli-guidelines&type=star&count=true&size=large" frameborder="0" scrolling="0" width="170" height="30" title="GitHub"></iframe>

ガイドや CLI デザインを議論したいなら [Discord に参加](https://discord.gg/EbAW5rUCkE)。

## 序文 {#foreword}

1980年代、パーソナルコンピュータに何かをさせたいなら、`C:\>` や `~$` の前で何をタイプすべきか知っている必要があった。
助けは分厚いスパイラル製本のマニュアル。
エラーメッセージは不透明。
Stack Overflow もなかった。
だが運よくインターネットにアクセスできれば、Usenet という初期インターネットのコミュニティから助けを得られた。
同じように苛立っている人々が集まる場所だ。
彼らは問題解決を手伝うか、少なくとも精神的支援と連帯感をくれた。

40年後、コンピュータははるかに身近になったが、低レベルのエンドユーザー制御はしばしば犠牲になった。
多くのデバイスではコマンドラインへのアクセス自体がない。
囲い込みやアプリストアという企業利益に反するためでもある。

今日のほとんどの人はコマンドラインが何かすら知らず、ましてやなぜ気にする必要があるのかも知らない。
計算機の先駆者 Alan Kay は [2017年のインタビュー](https://www.fastcompany.com/40435064/what-alan-kay-thinks-about-the-iphone-and-technology-now) でこう言った。「人々は計算とは何か理解していないから、iPhone の中にそれがあると思ってしまう。その錯覚は『Guitar Hero が本物のギターと同じ』という錯覚と同じくらい悪い」。

Kay の「本物のギター」は CLI ではない——少なくとも正確には。
彼が話していたのは、CLI の力を持ちながらテキストファイルでソフトを書くことを超える計算機のプログラミング方法だ。
Kay の弟子たちの間には、何十年も続いているテキスト中心の局所最適から抜け出す必要があるという信念がある。

コンピュータをまったく違う形でプログラムする未来を想像するのは刺激的だ。
今日でも、スプレッドシートは圧倒的に最も人気のあるプログラミング言語であり、ノーコード運動は有能なプログラマへの強い需要の一部を置き換えようとして急速に拡大している。

それでも、きしむような数十年もの制約と不可解な癖を抱えつつ、コマンドラインはなおコンピュータの最も _汎用的_ な領域だ。
カーテンを引いて裏側を見せ、何が本当に起きているのかを見せ、GUI では届かない精緻さと深さで機械と創造的に対話させる。
学びたい人ならほぼどのノートPCでも使える。
対話的にも、自動化にも使える。
そして、他のシステム要素ほど速くは変わらない。
その安定性には創造的価値がある。

だから、まだ持っているうちに、その有用性とアクセシビリティを最大化すべきだ。

初期の時代から、計算機をどうプログラムするかは大きく変わった。
昔のコマンドラインは _machine-first_ で、スクリプト基盤上の REPL 以上のものではなかった。
しかし汎用インタプリタ言語が花開くにつれ、シェルスクリプトの役割は縮小した。
今日のコマンドラインは _human-first_ だ。あらゆるツール、システム、プラットフォームへアクセスするテキストUIである。
以前はエディタが端末の中にあったが、今では端末がエディタの機能であることも多い。
`git` のようなマルチツールコマンドも増えた。
コマンドの中のコマンド、原子的機能ではなくワークフロー全体を実行する高レベルコマンド。

伝統的な UNIX 哲学に触発され、より楽しくアクセシブルな CLI 環境を促進したいという関心と、プログラマとしての経験に導かれ、コマンドライン・プログラムのベストプラクティスと設計原則を見直す時だと判断した。

コマンドライン万歳！

## はじめに {#introduction}

この文書は高レベルの設計思想と具体的なガイドラインの両方を扱う。
実務家としてあまり哲学しない方針なので、ガイドラインの比重が大きい。
例から学ぶのがよいと考え、多くの例を示した。

このガイドは emacs や vim のようなフルスクリーン端末プログラムは扱わない。
フルスクリーンはニッチで、その設計に携わる人はごく少ない。

このガイドは言語やツールにも中立だ。

このガイドの対象は？

- CLI プログラムを作っていて、UI 設計の原則や具体的ベストプラクティスが欲しい人。
- プロの「CLI UI デザイナー」なら素晴らしい。ぜひ学ばせてほしい。
- 40年の CLI 設計慣習に反する明白な失敗を避けたい人。
- 良い設計と親切なヘルプで人を喜ばせたい人。
- GUI プログラムを作っている人には向かない（ただし GUI のアンチパターンは学べるかもしれない）。
- Minecraft の没入型フルスクリーン CLI 移植を設計している人には向かない。
  （でもぜひ見てみたい！）

## 哲学 {#philosophy}

良い CLI 設計の基本原則だと考えるもの。

### 人間優先の設計 {#human-first-design}

伝統的に UNIX コマンドは、主に他のプログラムから使われる前提で書かれていた。
グラフィカルなアプリよりも、プログラミング言語の関数に近い。

今日、多くの CLI は人間が主に（あるいは専ら）使うのに、相互作用設計は過去の荷物を背負っている。
人間が主に使うなら、人間優先で設計すべき。

### 単純な部品を組み合わせる {#simple-parts-that-work-together}

[オリジナルの UNIX 哲学](https://en.wikipedia.org/wiki/Unix_philosophy)の核心は、小さく単純でクリーンなインターフェースのプログラムを組み合わせて大きなシステムを作るという考え方だ。
機能を詰め込むのではなく、必要に応じて組み替えられるほど十分にモジュール化する。

昔はパイプとシェルスクリプトがプログラム合成の要だった。
汎用インタプリタ言語の台頭で役割は減ったかもしれないが、消えたわけではない。
さらに CI/CD、オーケストレーション、構成管理といった大規模自動化が発展した。
合成可能性は今も重要。

幸い、この目的のために設計された UNIX 環境の長い慣習は今も役立つ。
標準入出力/標準エラー、シグナル、終了コードなどの仕組みがプログラム同士をうまく噛み合わせる。
行ベースのプレーンテキストはコマンド間のパイプが容易だ。
JSON はより構造化でき、コマンドラインツールと Web の統合を容易にする。

どんなソフトでも、人は想定外の使い方をする。
あなたのソフトは _必ず_ 大きなシステムの一部になる——選べるのは行儀の良い部品になるかどうかだけ。

重要なのは、合成可能性の設計と人間優先の設計は対立しないこと。
この文書の多くは、その両立の方法だ。

### プログラム間の一貫性 {#consistency-across-programs}

端末の慣習は指に染み付いている。
構文、フラグ、環境変数などを学ぶ初期コストはかかるが、プログラムが一貫していれば長期的に効率が上がる。

可能な限り、既存パターンに従うべきだ。
それが CLI を直感的で推測可能にし、ユーザーを効率的にする。

ただし一貫性が使いやすさと衝突することもある。
例えば、多くの古い UNIX コマンドはデフォルトでほとんど情報を出さず、慣れていない人には混乱や不安の原因になる。

慣習に従うことで可用性が損なわれるなら、破る判断もあり得る。
ただし慎重に。

### （ちょうど）十分に語る {#saying-just-enough}

端末は純粋な情報の世界だ。
情報そのものがインターフェースだと言うこともできるし、どのインターフェースにも情報が多すぎる/少なすぎる問題がある。

コマンドが数分ハングしてユーザーが壊れたかと疑い始めるなら「語りが少なすぎる」。
デバッグ出力を何ページも吐き、重要な情報がゴミの海に溺れるなら「語りが多すぎる」。
結果は同じ。明瞭さが欠け、ユーザーは混乱し苛立つ。

このバランスは難しいが、ユーザーを力づけ、奉仕するには不可欠だ。

### 発見容易性 {#ease-of-discovery}

機能の発見性では GUI が有利。
画面にすべてが見えているので、学習せずに探せるし、未知の機能も見つかる。

CLI はその逆だと見なされがちで、すべて覚える必要があるとされる。
1987年の [Macintosh Human Interface Guidelines](https://archive.org/details/applehumaninterf00appl) も「覚えるより見て指す (See-and-point)」を推奨している。

これは両立できる。
CLI の効率はコマンドを覚えることから来るが、コマンドが学習と記憶を助けてもよい。

発見可能な CLI は包括的なヘルプ、豊富な例、次に実行すべきコマンドの提案、エラー時の次の手の提案を提供する。
GUI から盗めるアイデアは多く、パワーユーザーにとっても学習と利用がしやすくなる。

_引用: The Design of Everyday Things (Don Norman), Macintosh Human Interface Guidelines_

### 会話を標準とする {#conversation-as-the-norm}

GUI 設計、特に初期は _メタファー_ を多用した。デスクトップ、ファイル、フォルダ、ゴミ箱。
コンピュータが正当性を確立しようとしていた時代には理にかなっていた。
メタファーの実装容易性は GUI の大きな利点だった。
しかし皮肉にも、CLI はずっと偶然のメタファーを体現している。CLI は会話だ。

最も単純なコマンドを超えると、プログラムの実行はたいてい一回では終わらない。
最初はうまくいかず、エラーを見て修正し、別のエラーを見て…を繰り返して成功する。
この反復失敗による学習は、ユーザーがプログラムと交わす会話のようだ。

試行錯誤だけが会話ではない。他にもある。

- ツールをセットアップするために一つ実行し、実際に使い始めるためのコマンドを学ぶ。
- 複数コマンドで準備し、最後のコマンドで実行（例: 複数の `git add` の後に `git commit`）。
- システム探索。`cd` と `ls` でディレクトリ構造を把握したり、`git log` と `git show` でファイルの履歴を辿る。
- 複雑な操作の本番前にドライランする。

会話として捉えると設計手法が見えてくる。
入力が無効なら修正候補を提示できる。
多段階の途中状態を明確にできる。
怖い操作の前に確認し、安心させられる。

ユーザーは意図の有無にかかわらずソフトと会話している。
最悪は敵対的な会話で、ユーザーを馬鹿にされた気分にさせ、恨みを生む。
最良は心地よい対話で、知識と達成感を与え、先へ進めてくれる。

_参考: [The Anti-Mac User Interface (Don Gentner and Jakob Nielsen)](https://www.nngroup.com/articles/anti-mac-interface/)_

### 堅牢性 {#robustness-principle}

堅牢性は客観的でもあり主観的でもある。
ソフトは当然 _堅牢であるべき_ で、予期しない入力を穏当に扱い、可能なら操作を冪等にすべきだ。
しかし _堅牢に感じる_ ことも重要だ。

壊れそうに感じさせたくない。
大きな機械のように即応し、軽いプラスチックの「ソフトスイッチ」ではないと感じさせたい。

主観的堅牢性には細部への注意と、何が起こり得るかを深く考えることが必要だ。
小さな積み重ねで、今何が起きているかを知らせ、よくあるエラーの意味を説明し、怖いスタックトレースを出さない。

一般に、堅牢性はシンプルさからも生まれる。
特例や複雑なコードはプログラムを脆くしがちだ。

### 共感 {#empathy}

コマンドラインツールはプログラマの創造的ツールキットなので、使って楽しいべきだ。
これはゲーム化や絵文字乱用を意味しない（絵文字自体が悪いわけではない 😉）。
ユーザーの味方で、成功してほしいと思い、問題と解決をよく考えたと感じさせることだ。

そう感じさせる確実な行動一覧はないが、この助言はその助けになるはずだ。
ユーザーを喜ばせるとは、常に期待を超えること。共感から始まる。

### カオス {#chaos}

端末の世界は混沌。
不一致は至る所にあり、手を止めさせ、迷いを生む。

それでもこの混沌が力の源であることは否定できない。
端末、そして UNIX 系の計算機環境は、構築できるものへの制約が少ない。
その空間であらゆる発明が花開いた。

この文書が既存パターンに従えと促しつつ、数十年の伝統に反する助言も並べるのは皮肉だ。
私たちも同じようにルールを破ってきた。

あなたもいつかルールを破る時が来る。
意図と目的を明確にして行え。

> “Abandon a standard when it is demonstrably harmful to productivity or user satisfaction.” — Jef Raskin, [The Humane Interface](https://en.wikipedia.org/wiki/The_Humane_Interface)

## ガイドライン {#guidelines}

コマンドライン・プログラムを良くするために具体的にできることの集まり。

最初のセクションは必須の項目。
これを間違えると、使いにくいか、CLI 市民としての行儀が悪いプログラムになる。

残りは「あると良い」項目。
時間と余裕があれば、平均的なプログラムよりずっと良くなる。

狙いは、設計についてあまり考えたくないならこのルールに従えばだいたい良いものになるということ。
一方で、考えた末にそのルールが自分のプログラムに合わないと判断したなら、それでも構わない。
（任意のルールに従わなかったからといって拒絶する中央権威はない。）

また、これらのルールは石に刻まれているわけではない。
良い理由で一般ルールに反対なら、[変更提案](https://github.com/cli-guidelines/cli-guidelines) をしてほしい。

### 基本 {#the-basics}

いくつかの基本ルールがある。
これを間違えると、プログラムは非常に使いにくいか、完全に壊れる。

**可能な限りコマンドライン引数パーサのライブラリを使う。**
言語の組み込みか良質なサードパーティを使う。
通常、引数、フラグ解析、ヘルプテキスト、スペルの提案まで扱ってくれる。

おすすめの例:

- マルチプラットフォーム: [docopt](http://docopt.org)
- Bash: [argbash](https://argbash.dev)
- Go: [Cobra](https://github.com/spf13/cobra), [cli](https://github.com/urfave/cli)
- Haskell: [optparse-applicative](https://hackage.haskell.org/package/optparse-applicative)
- Java: [picocli](https://picocli.info/)
- Julia: [ArgParse.jl](https://github.com/carlobaldassi/ArgParse.jl), [Comonicon.jl](https://github.com/comonicon/Comonicon.jl)
- Kotlin: [clikt](https://ajalt.github.io/clikt/)
- Node: [oclif](https://oclif.io/)
- Deno: [parseArgs](https://jsr.io/@std/cli/doc/parse-args/~/parseArgs)
- Perl: [Getopt::Long](https://metacpan.org/pod/Getopt::Long)
- PHP: [console](https://github.com/symfony/console), [CLImate](https://climate.thephpleague.com)
- Python: [Argparse](https://docs.python.org/3/library/argparse.html), [Click](https://click.palletsprojects.com/), [Typer](https://github.com/tiangolo/typer)
- Ruby: [TTY](https://ttytoolkit.org/)
- Rust: [clap](https://docs.rs/clap)
- Swift: [swift-argument-parser](https://github.com/apple/swift-argument-parser)

**成功時は 0、失敗時は非 0 の終了コードを返す。**
終了コードはスクリプトが成功/失敗を判断する方法なので、正しく報告する。
非 0 の終了コードは主要な失敗モードに割り当てる。

**出力は `stdout` へ。**
コマンドの主な出力は `stdout` に流す。
機械可読な出力も `stdout` に流すべきだ——パイプは通常ここを使う。

**メッセージは `stderr` へ。**
ログ、エラーなどは `stderr` に送る。
コマンドがパイプで連結されているとき、これらのメッセージはユーザーに表示され、次のコマンドには渡らない。

### ヘルプ {#help}

**求められたら詳細なヘルプを表示する。**
`-h` または `--help` が渡されたときにヘルプを表示する。
サブコマンドにも同様に適用する。

**デフォルトは簡潔なヘルプ。**
`myapp` または `myapp subcommand` が引数を必要とする場合、引数なしで実行されたら簡潔なヘルプを表示する。

ただし、プログラムがデフォルトで対話的（例: `npm init`）なら、このガイドラインは無視してよい。

簡潔なヘルプに含めるものは以下だけ:

- プログラムの説明。
- 1〜2個の例。
- フラグの説明（数が多い場合は除外）。
- 詳細は `--help` を渡すよう指示。

`jq` は良い例。
`jq` と入力すると、概要と例を表示し、`jq --help` を促す:

```
$ jq
jq - commandline JSON processor [version 1.6]

Usage:    jq [options] <jq filter> [file...]
    jq [options] --args <jq filter> [strings...]
    jq [options] --jsonargs <jq filter> [JSON_TEXTS...]

jq is a tool for processing JSON inputs, applying the given filter to
its JSON text inputs and producing the filter's results as JSON on
standard output.

The simplest filter is ., which copies jq's input to its output
unmodified (except for formatting, but note that IEEE754 is used
for number representation internally, with all that that implies).

For more advanced filters see the jq(1) manpage ("man jq")
and/or https://stedolan.github.io/jq

Example:

    $ echo '{"foo": 0}' | jq .
    {
        "foo": 0
    }

For a listing of options, use jq --help.
```

**`-h` と `--help` が渡されたら必ずヘルプを表示する。**
以下はすべてヘルプを表示すべき:

```
$ myapp
$ myapp --help
$ myapp -h
```

他のフラグや引数は無視してよい——どんなコマンドの末尾に `-h` を足してもヘルプが出るべき。
`-h` を別用途に使わない。

プログラムが `git` 風なら、以下もヘルプにする:

```
$ myapp help
$ myapp help subcommand
$ myapp subcommand --help
$ myapp subcommand -h
```

**フィードバック/課題の導線を用意する。**
トップレベルのヘルプに Web サイトや GitHub リンクを載せるのが一般的。

**ヘルプテキストから Web 版ドキュメントへリンクする。**
サブコマンド専用のページやアンカーがあれば直リンクする。
Web 側に詳細説明や補足がある場合に特に有用。

**例を先に出す。**
ユーザーは他の文書より例を使うので、ヘルプページの先頭に置く。特に複雑で一般的な用途は先に。
可能なら実際の出力も見せる。

例を連ねてストーリーとして見せ、複雑な使い方へ段階的に導くこともできる。

<!-- TK example? -->

**例が多すぎるなら別に置く。** チートシートコマンドや Web ページへ。
詳細な上級例は有用だが、ヘルプテキストを長くしすぎない。

他ツール連携などの複雑なユースケースには、完全なチュートリアルが適切なこともある。

**よく使うフラグ/コマンドをヘルプ冒頭に置く。**
フラグが多くても、よく使うものは先に表示する。
例えば Git は開始系やよく使うサブコマンドを先に出す:

```
$ git
usage: git [--version] [--help] [-C <path>] [-c <name>=<value>]
           [--exec-path[=<path>]] [--html-path] [--man-path] [--info-path]
           [-p | --paginate | -P | --no-pager] [--no-replace-objects] [--bare]
           [--git-dir=<path>] [--work-tree=<path>] [--namespace=<name>]
           <command> [<args>]

These are common Git commands used in various situations:

start a working area (see also: git help tutorial)
   clone      Clone a repository into a new directory
   init       Create an empty Git repository or reinitialize an existing one

work on the current change (see also: git help everyday)
   add        Add file contents to the index
   mv         Move or rename a file, a directory, or a symlink
   reset      Reset current HEAD to the specified state
   rm         Remove files from the working tree and from the index

examine the history and state (see also: git help revisions)
   bisect     Use binary search to find the commit that introduced a bug
   grep       Print lines matching a pattern
   log        Show commit logs
   show       Show various types of objects
   status     Show the working tree status
…
```

**ヘルプテキストに書式を使う。**
太字の見出しはスキャンしやすい。
ただし端末依存にならないようにし、エスケープ文字の壁を見せない。

<pre>
<code>
<strong>$ heroku apps --help</strong>
list your apps

<strong>USAGE</strong>
  $ heroku apps

<strong>OPTIONS</strong>
  -A, --all          include apps in all teams
  -p, --personal     list apps in personal account when a default team is set
  -s, --space=space  filter by space
  -t, --team=team    team to use
  --json             output in json format

<strong>EXAMPLES</strong>
  $ heroku apps
  === My Apps
  example
  example2

  === Collaborated Apps
  theirapp   other@owner.name

<strong>COMMANDS</strong>
  apps:create     creates a new app
  apps:destroy    permanently destroy an app
  apps:errors     view app errors
  apps:favorites  list favorited apps
  apps:info       show detailed app information
  apps:join       add yourself to a team app
  apps:leave      remove yourself from a team app
  apps:lock       prevent team members from joining an app
  apps:open       open the app in a web browser
  apps:rename     rename an app
  apps:stacks     show the list of available stacks
  apps:transfer   transfer applications to another user or team
  apps:unlock     unlock an app so any team member can join
</code>
</pre>

注: `heroku apps --help` を pager に通すと、エスケープ文字は出ない。

**ユーザーが間違えた時、推測できるなら提案する。**
例えば `brew update jq` なら `brew upgrade jq` を提案する。

提案コマンドを実行したいか聞いてもよいが、強制しない。
例:

```
$ heroku pss
 ›   Warning: pss is not a heroku command.
Did you mean ps? [y/n]:
```

修正した構文を勝手に実行したくなるかもしれないが、常に正しいとは限らない。

まず、無効入力は単なるタイポとは限らず、論理的な誤りやシェル変数の誤用のことも多い。
意図を決めつけるのは危険で、特に状態を変更する操作ならなおさら。

次に、ユーザー入力を書き換えると正しい構文を学べない。
事実上、その入力を正当な構文として永続的にサポートすることになる。
その判断は意図的に行い、両方の構文を文書化すること。

_参考: [“Do What I Mean”](http://www.catb.org/~esr/jargon/html/D/DWIM.html)_

**コマンドがパイプ入力を期待しているのに `stdin` が対話端末なら、すぐヘルプを出して終了する。**
`cat` のように黙って待たない。
代わりに `stderr` にログメッセージを出してもよい。

### ドキュメント {#documentation}

[ヘルプテキスト](#help) の目的は、ツールの概要、利用可能なオプション、最も一般的なタスクのやり方を短く伝えること。
一方でドキュメントは詳細を語る場所だ。
ツールが何のためにあり、何 _ではない_ のか、どう動き、どう使えば良いのかを説明する。

**Web ベースのドキュメントを提供する。**
オンライン検索でき、他人に特定箇所をリンクできる必要がある。
Web は最も包括的なドキュメント形式だ。

**端末ベースのドキュメントも提供する。**
端末ドキュメントはアクセスが速く、インストール済みバージョンと同期し、ネットなしでも読める。

**man ページの提供を検討する。**
[man ページ](https://en.wikipedia.org/wiki/Man_page)は UNIX の元祖ドキュメントで、今も使われる。
多くのユーザーは `man mycmd` を最初に試す。
生成を容易にするなら [ronn](http://rtomayko.github.io/ronn/ronn.1.html) などを使う（Web ドキュメント生成にも使える）。

ただし `man` を知らない人もいるし、全プラットフォームで動くわけではない。
そのため、端末ドキュメントはツール自身からもアクセスできるようにする。
例えば `git` や `npm` は `help` サブコマンドで man ページにアクセスでき、`npm help ls` は `man npm-ls` と同等。

```
NPM-LS(1)                                                            NPM-LS(1)

NAME
       npm-ls - List installed packages

SYNOPSIS
         npm ls [[<@scope>/]<pkg> ...]

         aliases: list, la, ll

DESCRIPTION
       This command will print to stdout all the versions of packages that are
       installed, as well as their dependencies, in a tree-structure.

       ...
```

### 出力 {#output}

**人間可読な出力が最重要。**
人間が先、機械は後。
特定の出力ストリーム（`stdout`/`stderr`）が人間に読まれているかの最も簡単な判定は _TTY かどうか_ だ。
どの言語にも判定ユーティリティ/ライブラリがある（例: [Python](https://stackoverflow.com/questions/858623/how-to-recognize-whether-a-script-is-running-on-a-tty), [Node](https://nodejs.org/api/process.html#process_a_note_on_process_i_o), [Go](https://github.com/mattn/go-isatty)）。

_TTY とは何か: [参考](https://unix.stackexchange.com/a/4132)_

**ユーザビリティを損なわない範囲で機械可読な出力を用意する。**
UNIX における普遍的インターフェースはテキストのストリームだ。
プログラムは通常テキスト行を出力し、通常テキスト行を入力として期待する。
ゆえに複数プログラムを合成できる。
これはスクリプト作成のためだけでなく、人間の使いやすさにも寄与する。
例えば、ユーザーが出力を `grep` にパイプして期待通りに動くべきだ。

> “Expect the output of every program to become the input to another, as yet unknown, program.”
> — [Doug McIlroy](http://web.archive.org/web/20220609080931/https://homepage.cs.uri.edu/~thenry/resources/unix_art/ch01s06.html)

**人間可読な出力が機械可読性を壊すなら `--plain` を用意する。**
`grep` や `awk` と連携できるよう、プレーンな表形式で出力する。

<!-- (TK example with and without --plain) -->

例えば行ベースのテーブル表示で、セルを複数行に分割して画面幅に収める場合。
これは「1行=1レコード」の前提を崩すので、`--plain` フラグで操作を無効化し、1行1レコードを出力すべき。

**`--json` が渡されたら整形 JSON で出力する。**
JSON はプレーンテキストより構造化でき、複雑なデータ構造の出力や処理が容易。
[`jq`](https://stedolan.github.io/jq/) はコマンドライン JSON ツールとして一般的で、JSON の出力/操作ツールの[エコシステム](https://ilya-sher.org/2018/04/10/list-of-json-tools-for-command-line/)もある。

JSON は Web でも広く使われるため、`curl` を使って Web サービスへ直接パイプ入出力できる。

**成功時の出力は出すが短くする。**
伝統的に UNIX コマンドは問題がなければ出力しない。
スクリプトでは理にかなうが、人間が使うとハングや故障に見えることがある。
例えば `cp` は時間がかかっても何も表示しない。

何も出さないのが最良というケースは稀で、通常は「少なめ」に倒すのが良い。

シェルスクリプト向けに出力を無くしたい場合、`stderr` を `/dev/null` に捨てるのは不格好なので、非必須出力を抑える `-q` オプションを提供するとよい。

**状態を変更したらユーザーに伝える。**
システムの状態を変えるコマンドでは、何が起きたかを説明する価値が高い。
結果がユーザーの要求と直接一致しない場合ほど重要。

例えば `git push` は何をしているか、リモートの新しい状態が何かを明確に示す:

```
$ git push
Enumerating objects: 18, done.
Counting objects: 100% (18/18), done.
Delta compression using up to 8 threads
Compressing objects: 100% (10/10), done.
Writing objects: 100% (10/10), 2.09 KiB | 2.09 MiB/s, done.
Total 10 (delta 8), reused 0 (delta 0), pack-reused 0
remote: Resolving deltas: 100% (8/8), completed with 8 local objects.
To github.com:replicate/replicate.git
 + 6c22c90...a2a5217 bfirsh/fix-delete -> bfirsh/fix-delete
```

**現在の状態を見やすくする。**
プログラムが複雑な状態変更をするのに、ファイルシステムから直感的に分からない場合は、状態の可視化を容易にする。

例えば `git status` は現在の状態を可能な限り多く伝え、状態を変えるためのヒントも出す:

```
$ git status
On branch bfirsh/fix-delete
Your branch is up to date with 'origin/bfirsh/fix-delete'.

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   cli/pkg/cli/rm.go

no changes added to commit (use "git add" and/or "git commit -a")
```

**次に実行すべきコマンドを提案する。**
複数コマンドで構成されるワークフローでは、次に実行すべきコマンドを提案すると学習や新機能発見に役立つ。
上の `git status` も状態を変更するコマンドを提案している。

**プログラム内部の境界を越える操作は明示的にする。**
例:

- ユーザーが明示的に渡していないファイルを読み書きする（キャッシュなど内部状態を保存するファイルは除く）。
- リモートサーバーへ通信しファイルをダウンロードする。

**ASCII アートで情報密度を上げる。**
例えば `ls` は権限をスキャンしやすく表示する。
最初はほとんど無視でき、慣れるにつれて多くのパターンを拾える。

```
-rw-r--r-- 1 root root     68 Aug 22 23:20 resolv.conf
lrwxrwxrwx 1 root root     13 Mar 14 20:24 rmt -> /usr/sbin/rmt
drwxr-xr-x 4 root root   4.0K Jul 20 14:51 security
drwxr-xr-x 2 root root   4.0K Jul 20 14:53 selinux
-rw-r----- 1 root shadow  501 Jul 20 14:44 shadow
-rw-r--r-- 1 root root    116 Jul 20 14:43 shells
drwxr-xr-x 2 root root   4.0K Jul 20 14:57 skel
-rw-r--r-- 1 root root      0 Jul 20 14:43 subgid
-rw-r--r-- 1 root root      0 Jul 20 14:43 subuid
```

**色は意図的に使う。**
ユーザーの注意を引きたい文字をハイライトしたり、エラーを赤で示したり。
使い過ぎは禁物——すべてが色だと色の意味が失われ読みづらくなる。

**端末でない場合やユーザーが要求した場合は色を無効化する。**
次の場合は色を無効にする:

- `stdout` または `stderr` が対話端末 (TTY) ではない。
  個別に判定するのが良い——`stdout` をパイプしていても `stderr` には色が有用な場合がある。
- `NO_COLOR` 環境変数が設定され空でない（値に関わらず）。
- `TERM` 環境変数が `dumb`。
- ユーザーが `--no-color` を渡した。
- プログラム専用の `MYAPP_NO_COLOR` を用意してもよい。

_参考: [no-color.org](https://no-color.org/), [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)_

**`stdout` が対話端末でないならアニメーションを表示しない。**
CI ログの進捗バーがクリスマスツリー化するのを防ぐ。

**明確化に役立つなら記号や絵文字を使う。**
いくつかの要素を区別したいとき、注意を引きたいとき、少し個性を足したいとき、文字より絵の方が伝わることがある。
ただしやりすぎると散らかって見えたり、玩具のように感じられる。

例えば [yubikey-agent](https://github.com/FiloSottile/yubikey-agent) は絵文字で構造を付け、❌ で重要な情報に注意を向けさせている:

```shell-session
$ yubikey-agent -setup
🔐 The PIN is up to 8 numbers, letters, or symbols. Not just numbers!
❌ The key will be lost if the PIN and PUK are locked after 3 incorrect tries.

Choose a new PIN/PUK:
Repeat the PIN/PUK:

🧪 Reticulating splines …

✅ Done! This YubiKey is secured and ready to go.
🤏 When the YubiKey blinks, touch it to authorize the login.

🔑 Here's your new shiny SSH public key:
ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBCEJ/
UwlHnUFXgENO3ifPZd8zoSKMxESxxot4tMgvfXjmRp5G3BGrAnonncE7Aj11pn3SSYgEcrrn2sMyLGpVS0=

💭 Remember: everything breaks, have a backup plan for when this YubiKey does.
```

**デフォルトで、開発者だけが理解できる情報は出力しない。**
出力が開発者の理解のためだけなら、通常ユーザーには不要で、デフォルトでは出すべきではない。必要なら verbose で。

外部の人、プロジェクトに新しい人からのユーザビリティ・フィードバックを歓迎する。
コードに近すぎると重要な問題が見えないので助けになる。

**`stderr` をログファイル扱いしない（少なくともデフォルトでは）。**
`ERR` や `WARN` などのログレベルや余計な文脈情報は出さない。必要なら verbose で。

**大量のテキストは pager（例: `less`）を使う。**
例えば `git diff` はデフォルトで pager を使う。
pager は実装を間違えると体験が悪化するので注意。
`stdin` または `stdout` が対話端末のときだけ pager を使う。

`less` の適切なオプション例は `less -FIRX`。
1画面ならページングせず、検索は大文字小文字を無視し、色と書式を有効化し、終了後も内容を残す。

言語によっては `less` へのパイプより堅牢なライブラリがある。
例えば Python の [pypager](https://github.com/prompt-toolkit/pypager)。

### エラー {#errors}

ドキュメントを見る最も一般的な理由はエラー解決だ。
エラー自体をドキュメントにできれば、ユーザーの時間を大幅に節約できる。

**エラーを捕捉し、人間向けに書き換える。**
予期できるエラーは捕捉し、役に立つメッセージに書き換える。
会話のように、ユーザーが間違えた時に正しい方向へ導く。
例: 「file.txt に書き込めません。`chmod +w file.txt` で書き込み可能にする必要があるかもしれません。」

**S/N 比が重要。**
無関係な出力が多いほど、ユーザーが問題を理解するのに時間がかかる。
同種のエラーが複数あるなら、似た行を大量に出す代わりに 1 つの説明ヘッダにまとめることを検討する。

**ユーザーが最初に見る場所を考える。**
最重要情報は出力の末尾に置く。
赤いテキストは目に付くので意図的かつ控えめに使う。

**予期しない/説明不能なエラーにはデバッグとトレース情報、バグ報告方法を提供する。**
ただし S/N 比を忘れない。理解できない情報で圧倒しない。
デバッグログは端末に出す代わりにファイルへ書くのもよい。

**バグ報告を手間なくする。**
URL を提供し、可能な限り情報を事前入力するのは良い。

_参考: [Google: Writing Helpful Error Messages](https://developers.google.com/tech-writing/error-messages), [Nielsen Norman Group: Error-Message Guidelines](https://www.nngroup.com/articles/error-message-guidelines)_

### 引数とフラグ {#arguments-and-flags}

用語メモ:

- _Arguments_（args）はコマンドの位置引数。
  例: `cp` に渡すファイルパス。
  args の順序は重要で、`cp foo bar` は `cp bar foo` と意味が違う。
- _Flags_ は名前付きパラメータで、ハイフン+1文字（`-r`）またはダブルハイフン+複数文字（`--recursive`）。
  ユーザー指定の値を含む場合もある（`--file foo.txt` または `--file=foo.txt`）。
  一般にフラグの順序は意味に影響しない。

**args よりフラグを優先する。**
少しタイプは増えるが、何が起きているかが明確になる。
将来的な入力変更も容易。
args だけだと新しい入力を追加できず、互換性を壊す/曖昧になることがある。

_引用: [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)._

**すべてのフラグにフルネーム版を用意する。**
例: `-h` と `--help`。
スクリプトでは長いほうが読みやすく、意味を調べる手間が減る。

_引用: [GNU Coding Standards](https://www.gnu.org/prep/standards/html_node/Command_002dLine-Interfaces.html)._

**1文字フラグはよく使うものだけにする。**
特にサブコマンド構成のトップレベルでは重要。
短い名前空間を汚すと、将来追加するフラグで苦労する。

**単純な複数ファイル操作なら複数の引数は OK。**
例: `rm file1.txt file2.txt file3.txt`。
グロブにも対応できる: `rm *.txt`。

**異なる意味の引数が 2 つ以上あるなら、たぶん設計が悪い。**
例外は一般的で主要な操作で、短さが覚える価値に勝るとき。
例: `cp <source> <destination>`。

_引用: [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)._

**標準があるなら標準のフラグ名を使う。**
一般的なコマンドが使うフラグ名は踏襲すべき。
ユーザーが複数のオプションを覚えずに済み、ヘルプを見なくても推測できる。

よく使われるオプション一覧:

- `-a`, `--all`: すべて。
  例: `ps`, `fetchmail`.
- `-d`, `--debug`: デバッグ出力。
- `-f`, `--force`: 強制。
  例: `rm -f` は権限がなくても削除を強行する。
  破壊的操作で通常は確認が必要な場合に、スクリプトで強制するためにも使える。
- `--json`: JSON 出力。
  [出力](#output) 参照。
- `-h`, `--help`: ヘルプ。
  ヘルプ以外の意味に使わない。
  [ヘルプ](#help) 参照。
- `-n`, `--dry-run`: ドライラン。
  コマンドを実行せず、実行した場合に起きる変更を説明する。
  例: `rsync`, `git add`.
- `--no-input`: [対話](#interactivity) 参照。
- `-o`, `--output`: 出力ファイル。
  例: `sort`, `gcc`.
- `-p`, `--port`: ポート。
  例: `psql`, `ssh`.
- `-q`, `--quiet`: 静かに。
  出力を減らす。
  人間向けの出力をスクリプトでは隠したい時に便利。
- `-u`, `--user`: ユーザー。
  例: `ps`, `ssh`.
- `--version`: バージョン。
- `-v`: verbose か version のどちらかを意味しがち。
  verbose は `-d` にし、`-v` を version にするか、混乱を避けるため使わない手もある。

**デフォルトを大多数にとって正しいものにする。**
設定可能なのは良いが、多くのユーザーは適切なフラグを見つけて毎回使わない（エイリアスもしない）。
デフォルトでないと、多くのユーザー体験が悪化する。

例えば `ls` は歴史的理由でデフォルトが簡素だが、今日設計するなら `ls -lhF` がデフォルトになりそうだ。

**ユーザー入力を促す。**
引数やフラグが渡されなければ、入力を促す。
（[対話](#interactivity) も参照）

**プロンプトを _必須_ にしない。**
必ずフラグ/引数で入力できるようにする。
`stdin` が対話端末でないならプロンプトをスキップし、必要なフラグ/引数を要求する。

**危険な操作の前に確認する。**
一般的には、対話時は `y`/`yes` を入力させるか、非対話時は `-f`/`--force` を要求する。

「危険」は主観で、段階がある:

- **軽度:** 小さなローカル変更（例: ファイル削除）。
  確認する場合も、しない場合もある。
  例えばコマンド名が明確に「delete」のようなら、確認不要かもしれない。
- **中程度:** ディレクトリ削除のような大きなローカル変更、リモート資源削除、容易に取り消せない複雑な一括変更。
  通常は確認すべき。
  ドライランで事前に内容を見せることを検討する。
- **重大:** リモートアプリやサーバー全体の削除など。
  確認だけでなく、誤操作を避けるために確認しづらくする。
  削除対象の名前の入力など、軽くない文字列を要求する。
  それでもスクリプト化できるよう、`--confirm="name-of-thing"` のようなフラグを用意する。

非自明な破壊の可能性も考える。
例えば、設定ファイルの数値を 10 から 1 に変えると 9 件が暗黙削除されるような場合は重大リスクとして扱い、誤操作しづらくすべきだ。

**入出力がファイルなら `-` を `stdin`/`stdout` に対応させる。**
他コマンドの出力を自コマンドの入力に、逆も可能になり、一時ファイルが不要。
例: `tar` は `stdin` から展開できる:

```
$ curl https://example.com/something.tar.gz | tar xvf -
```

**フラグが任意値を取るなら “none” のような特別語を許可する。**
例: `ssh -F` は代替 `ssh_config` ファイルの任意値を取るが、`ssh -F none` で設定なしで動かせる。
空文字だけにしない。引数がフラグ値か通常引数か曖昧になる。

**可能なら引数・フラグ・サブコマンドの順序依存をなくす。**
多くの CLI、特にサブコマンド構成では、引数を置ける位置が暗黙的に決まっている。
例えば `--foo` がサブコマンド前でしか効かない:

```
$ mycmd --foo=1 subcmd
works

$ mycmd subcmd --foo=1
unknown flag: --foo
```

これは非常に混乱を招く。
特にユーザーがよくやるのは「前回のコマンドを上矢印で出し、末尾にオプションを足して再実行」だからだ。
可能なら両方を同等に扱う。
ただしパーサの制約に当たることもある。

**秘密情報をフラグから直接読まない。**
`--password` のようなフラグは `ps` 出力やシェル履歴に漏れる。
また、環境変数で秘密を扱う悪習も誘発する。
（環境変数は他ユーザーに読まれたり、デバッグログに載ったりしやすい。）

機密データは `--password-file` のようなファイル経由、または `stdin` で受け取るのが良い。
`--password-file` は多くの文脈で秘密を目立たず渡せる。

（Bash なら `--password $(< password.txt)` でファイル内容をフラグに入れられるが、上記と同じ問題があるので避ける。）

### 対話 {#interactivity}

**`stdin` が対話端末 (TTY) のときだけプロンプトや対話要素を使う。**
データをパイプしているのか、スクリプト内で実行されているのかを判定する信頼性の高い方法だ。
その場合プロンプトは使えないので、どのフラグを渡すべきかをエラーで伝える。

**`--no-input` が渡されたらプロンプトや対話を行わない。**
ユーザーが明示的に対話を無効化できる。
入力が必須なら失敗させ、フラグで渡す方法を教える。

**パスワード入力時は表示しない。**
端末のエコーをオフにする。
言語の補助機能を使う。

**ユーザーが抜けられるようにする。**
抜け方を明確にする。
（vim のようにしない。）
ネットワーク I/O などでハングしても Ctrl-C が必ず効くようにする。
SSH/tmux/telnet のように Ctrl-C で抜けられないラッパーなら、抜け方を明確にする。
例えば SSH は `~` エスケープシーケンスを持つ。

### サブコマンド

十分に複雑なツールなら、サブコマンドを使うことで複雑さを減らせる。
非常に密接に関連する複数ツールがあるなら、単一コマンドに統合すると使いやすく発見もしやすい（例: RCS vs. Git）。

サブコマンドは共有に向く——グローバルフラグ、ヘルプ、設定、ストレージ機構など。

**サブコマンド間で一貫性を保つ。**
同じ意味には同じフラグ名、出力フォーマットも揃える。

**複数レベルのサブコマンドは命名を揃える。**
複雑なソフトで多数のオブジェクトと操作がある場合、「名詞 + 動詞」2階層が一般的。
例: `docker container create`。
異なるオブジェクト間で動詞を揃える。

`名詞 動詞` でも `動詞 名詞` でも良いが、前者の方が一般的。

_参考: [User experience, CLIs, and breaking the world, by John Starich](https://uxdesign.cc/user-experience-clis-and-breaking-the-world-baed8709244f)._

**曖昧または似た名前のコマンドを避ける。**
例: “update” と “upgrade” の2つのサブコマンドは混乱を招く。
別の言葉にするか、追加語で区別する。

### 堅牢性 {#robustness-guidelines}

**ユーザー入力を検証する。**
ユーザー入力は必ずいつか壊れる。
早期にチェックして悪いことが起きる前に止め、[分かるエラー](#errors) にする。

**速さより応答性。**
100ms 以内にユーザーへ何か出す。
ネットワークリクエストをするなら、実行前に何か出してハングに見えないようにする。

**時間がかかるなら進捗を見せる。**
しばらく出力がないと壊れて見える。
良いスピナーや進捗表示は、実際より速く感じさせる。

Ubuntu 20.04 には端末下部に貼り付く良いプログレスバーがある。

<!-- (TK reproduce this as a code block or animated SVG) -->

進捗バーが長時間止まると、処理が続いているのかクラッシュしたのか分からない。
残り時間の推定や、少なくともアニメーションで「動作中」を示すと安心できる。

進捗バー生成には良いライブラリが多い。
例: Python の [tqdm](https://github.com/tqdm/tqdm)、Go の [schollz/progressbar](https://github.com/schollz/progressbar)、Node.js の [node-progress](https://github.com/visionmedia/node-progress)。

**並列化はできるなら行うが慎重に。**
シェルで進捗を出すのは難しい。並列ならさらに難しい。
堅牢で、出力が混ざって混乱しないようにする。
ライブラリが使えるなら使う——自分で書きたくない種類のコード。
Python の [tqdm](https://github.com/tqdm/tqdm) や Go の [schollz/progressbar](https://github.com/schollz/progressbar) は複数バーをネイティブに扱える。

利点は大きく、例えば `docker pull` の複数バーは状況把握に役立つ:

```
$ docker image pull ruby
Using default tag: latest
latest: Pulling from library/ruby
6c33745f49b4: Pull complete
ef072fc32a84: Extracting [================================================>  ]  7.569MB/7.812MB
c0afb8e68e0b: Download complete
d599c07d28e6: Download complete
f2ecc74db11a: Downloading [=======================>                           ]  89.11MB/192.3MB
3568445c8bf2: Download complete
b0efebc74f25: Downloading [===========================================>       ]  19.88MB/22.88MB
9cb1ba6838a0: Download complete
```

注意点: 進捗バーの裏にログを隠すと、うまくいっている時は理解しやすいが、エラー時はログを必ず表示する。
さもないとデバッグが非常に難しくなる。

**タイムアウトを設ける。**
ネットワークタイムアウトは設定可能にし、妥当なデフォルトを持たせ、永遠にハングしないようにする。

**回復可能にする。**
一時的な失敗（ネット接続が落ちた等）なら、`<up>` + `<enter>` で前回の続きから再開できるべき。

**クラッシュオンリーにする。**
冪等性の次の段階。
失敗や割り込み後に後片付けが不要、または次回に回せるなら、即時終了できる。
堅牢性と応答性の両方に効く。

_引用: [Crash-only software: More than meets the eye](https://lwn.net/Articles/191059/)._

**人はプログラムを誤用する。**
備えること。
スクリプトに包まれ、劣悪なネット環境で使われ、同時に多重実行され、未知の環境で動かされる。
（macOS のファイルシステムが大文字小文字を区別しないが保持はする、というのは知っている？）

### 将来互換 {#future-proofing}

ソフトウェアでは、長期かつ十分に文書化された非推奨プロセスなしにインターフェースを変えないことが重要。
サブコマンド、引数、フラグ、設定ファイル、環境変数はすべてインターフェースであり、維持する責任がある。
（[セマンティックバージョニング](https://semver.org/)でも、毎月メジャーアップするなら意味がない。）

**可能なら変更は追加的に。**
既存フラグの振る舞いを互換性なく変えるより、新しいフラグを追加する。
ただしインターフェースが膨らみすぎない範囲で。
（[フラグを args より優先](#arguments-and-flags) も参照。）

**非追加的変更の前に警告する。**
いずれ破壊的変更を避けられないことがある。
その前にプログラム内で警告する。非推奨にするフラグが渡されたら、もうすぐ変わると伝える。
今すぐ使い方を修正して将来互換にできる道を示し、方法を伝える。

可能なら、ユーザーが使い方を変えたことを検知し、警告を止める。
そうすれば変更を出してもユーザーは気づかない。

**人間向け出力の変更は通常 OK。**
インターフェースの改善には反復が必要で、出力がインターフェースだとすると改善できない。
スクリプトでは `--plain` や `--json` を使って安定させるよう促す（[出力](#output)参照）。

**万能サブコマンドを作らない。**
最も使われるサブコマンドを省略可能にしたくなることがある。
例えば任意のシェルコマンドを実行する `run`:

    $ mycmd run echo "hello world"

最初の引数が既存サブコマンド名でなければ `run` とみなすと、次のように短くできる:

    $ mycmd echo "hello world"

しかし重大な欠点がある。
`echo` というサブコマンドを将来追加できなくなる——_あらゆる_ サブコマンドが追加しづらくなる。
`mycmd echo` を使うスクリプトがあれば、アップグレード後に別の動作になる。

**サブコマンドの任意省略を許さない。**
例えば `install` サブコマンドがあり、非曖昧な接頭辞なら `mycmd ins` や `mycmd i` でも許可するとする。
これで `i` で始まるサブコマンドを将来追加できなくなる。
`i` が `install` を意味すると期待するスクリプトが存在するため。

エイリアス自体は悪くないが、明示的で安定しているべき。

**「時限爆弾」を作らない。**
20年後を想像してほしい。
今と同じように動くか、外部依存が変化/消滅して動かなくなるか？
20年後に存在しない可能性が高いサーバーは、今あなたが運用しているサーバーだ。
（ただし Google Analytics へのブロッキング呼び出しを埋め込むのも避ける。）

### シグナルと制御文字 {#signals}

**ユーザーが Ctrl-C（INT）を押したら、できるだけ早く終了する。**
クリーンアップ開始前に即座に何かを表示する。
クリーンアップにはタイムアウトを付け、永遠にハングしないようにする。

**時間のかかるクリーンアップ中に Ctrl-C が押されたらスキップする。**
2回目の Ctrl-C で何が起こるか（破壊的なら）を伝える。

例: Docker Compose を終了するとき、2回目の Ctrl-C で即停止する:

```
$  docker-compose up
…
^CGracefully stopping... (press Ctrl+C again to force)
```

プログラムはクリーンアップが走っていない状態で起動される可能性を想定すべき。
（[Crash-only software: More than meets the eye](https://lwn.net/Articles/191059/) 参照。）

### 設定 {#configuration}

コマンドラインツールには多様な設定タイプがあり、供給方法も多い（フラグ、環境変数、プロジェクト設定ファイル）。
各設定の最適な供給方法は、特に _特異性_・_安定性_・_複雑性_ に依存する。

設定は一般に以下のカテゴリに分かれる:

1.  コマンド実行ごとに変わりやすい。

    例:
    - デバッグ出力レベルの設定
    - セーフモードやドライランの有効化

    推奨: **[フラグ](#arguments-and-flags) を使う。**
    [環境変数](#environment-variables) も使える場合がある。

2.  実行ごとには概ね安定だが常に同じではない。
    プロジェクト間で変わる可能性がある。
    同じプロジェクト内でもユーザー間で違う。

    これは多くの場合、個々のコンピュータ固有の設定。

    例:
    - 起動に必要な項目の非デフォルトパス
    - 出力の色の有無
    - HTTP プロキシ設定

    推奨: **[フラグ](#arguments-and-flags) とおそらく [環境変数](#environment-variables) も使う。**
    ユーザーはシェルプロフィールに設定して全体適用したり、プロジェクトごとの `.env` に設定したりする。

    複雑なら専用設定ファイルにしてもよいが、通常は環境変数で十分。

3.  プロジェクト内で安定し、全ユーザーで同じ。

    バージョン管理に入るべき設定。
    `Makefile`、`package.json`、`docker-compose.yml` などが例。

    推奨: **コマンド専用のバージョン管理ファイルを使う。**

**XDG 仕様に従う。**
2010年に X Desktop Group（現 [freedesktop.org](https://freedesktop.org)）は設定ファイルのベースディレクトリの仕様を作った。
目的の一つはホームディレクトリの dotfile 増殖を抑え、`~/.config` を一般的に使うこと。
XDG Base Directory Specification（[仕様](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)、[要約](https://wiki.archlinux.org/index.php/XDG_Base_Directory#Specification)）は yarn, fish, wireshark, emacs, neovim, tmux など多くのプロジェクトで使われている。

**自分のプログラム以外の設定を自動変更するなら、必ず同意を取り、何をしているか正確に伝える。**
既存ファイルへの追記（例: `/etc/crontab`）より、新しい設定ファイル作成（例: `/etc/cron.d/myapp`）を優先。
システム全体の設定ファイルに追記/変更する必要があるなら、日付入りコメントで自分の変更箇所を区切る。

**設定の適用順序を守る。**
優先順位は以下（高い→低い）:

- フラグ
- 実行シェルの環境変数
- プロジェクトレベル設定（例: `.env`）
- ユーザーレベル設定
- システム全体設定

### 環境変数 {#environment-variables}

**環境変数は「実行コンテキストで変わる振る舞い」のためのもの。**
環境変数の「環境」とは端末セッション、つまりコマンド実行の文脈。
従って env var は実行ごとに変わるか、同一マシンのセッション間で変わるか、複数マシンのプロジェクト間で変わり得る。

環境変数はフラグや設定パラメータと重複する場合もあれば、独立する場合もある。
[設定](#configuration) で使い分けの指針を示す。

**可搬性最大化のため、環境変数名は大文字・数字・アンダースコアのみ（先頭は数字不可）。**
つまり `O_O` と `OWO` だけが有効な顔文字。

**環境変数の値は1行を目指す。**
複数行も可能だが `env` コマンドの扱いが悪化する。

**一般的に使われる名前を横取りしない。**
POSIX 標準の env var 一覧は[ここ](https://pubs.opengroup.org/onlinepubs/009695399/basedefs/xbd_chap08.html)。

**可能なら汎用環境変数を使う。**

- `NO_COLOR`: 色無効化（[出力](#output) 参照）。`FORCE_COLOR` は色強制。
- `DEBUG`: より冗長な出力。
- `EDITOR`: ファイル編集や複数行入力を促す場合。
- `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, `NO_PROXY`: ネットワーク操作がある場合。
  （HTTP ライブラリが既に参照している場合もある。）
- `SHELL`: ユーザーの好みのシェルで対話セッションを開く場合。
  （シェルスクリプト実行なら `/bin/sh` のような特定インタプリタを使う。）
- `TERM`, `TERMINFO`, `TERMCAP`: 端末依存のエスケープシーケンスを使う場合。
- `TMPDIR`: 一時ファイルを作る場合。
- `HOME`: 設定ファイルの場所を探す場合。
- `PAGER`: 出力をページングしたい場合。
- `LINES`, `COLUMNS`: 画面サイズに依存する出力（例: テーブル）。

**適切なら `.env` から環境変数を読む。**
特定ディレクトリでの作業中に変わりにくい環境変数があるなら、ローカル `.env` を読んでプロジェクトごとに設定できるようにする。
多くの言語に `.env` 読み込みライブラリがある（[Rust](https://crates.io/crates/dotenv), [Node](https://www.npmjs.com/package/dotenv), [Ruby](https://github.com/bkeepers/dotenv)）。

**`.env` を正式な[設定ファイル](#configuration)の代替にしない。**
`.env` には制約が多い:

- `.env` は一般にソース管理されない
- （よって設定履歴が残らない）
- データ型が文字列しかない
- 整理されにくい
- 文字エンコーディング問題を起こしやすい
- 機密情報や鍵素材が入りがちで、本来はより安全に保管すべき

これらが使いやすさや安全性を阻害しそうなら、専用設定ファイルのほうが適切。

**秘密情報を環境変数から読まない。**
環境変数は便利だが漏洩しやすいことが実証されている:

- export された環境変数は全プロセスへ渡り、ログに漏れたり流出しやすい
- `curl -H "Authorization: Bearer $BEARER_TOKEN"` のようなシェル展開はプロセス情報に漏れる
  （cURL は `-H @filename` でファイルから安全に読み込める）
- Docker コンテナの環境変数は `docker inspect` で閲覧できる
- systemd ユニットの環境変数は `systemctl show` で読める

秘密は資格情報ファイル、パイプ、`AF_UNIX` ソケット、秘密管理サービス、その他の IPC で受け取るべき。

### 命名 {#naming}

> “Note the obsessive use of abbreviations and avoidance of capital letters; [Unix] is a system invented by people to whom repetitive stress disorder is what black lung is to miners.
> Long names get worn down to three-letter nubbins, like stones smoothed by a river.”
> — Neal Stephenson, _[In the Beginning was the Command Line](https://web.stanford.edu/class/cs81n/command.txt)_

プログラム名は CLI では特に重要だ。ユーザーはそれを何度もタイプするので、覚えやすく打ちやすくあるべき。

**シンプルで覚えやすい単語にする。**
ただし一般的すぎると他コマンドと衝突し混乱する。
例: ImageMagick と Windows の両方が `convert` を使っていた。

**小文字のみを使い、必要ならダッシュを使う。**
`curl` は良いが `DownloadURL` は良くない。

**短くする。**
ユーザーは頻繁にタイプする。
ただし _短すぎる_ のは避ける。最短の名前は `cd`, `ls`, `ps` など日常的なユーティリティに温存すべき。

**打ちやすくする。**
一日中タイプされるなら、指に優しい名前にする。

実例: Docker Compose が `docker compose` になる前は [`plum`](https://github.com/aanand/fig/blob/0eb7d308615bae1ad4be1ca5112ac7b6b6cbfbaf/setup.py#L26) だった。
片手では厳しい綱渡りのような名前で、すぐに [`fig`](https://github.com/aanand/fig/commit/0cafdc9c6c19dab2ef2795979dc8b2f48f623379) に改名された。
短く、流れるように打てる。

_参考: [The Poetics of CLI Command Names](https://smallstep.com/blog/the-poetics-of-cli-command-names/)_

### 配布 {#distribution}

**可能なら単一バイナリで配布する。**
言語が標準でバイナリ生成しないなら、[PyInstaller](https://www.pyinstaller.org/) のようなものがないか検討する。
どうしても単一バイナリが無理なら、プラットフォーム標準のパッケージインストーラを使い、簡単に削除できないファイルを散乱させない。
ユーザーのコンピュータへの侵入は最小限に。

言語固有のツール（例: コードリンタ）ならこのルールは当てはまらない。
その言語のインタプリタが入っている前提でよい。

**アンインストールを簡単にする。**
手順が必要なら、インストール手順の末尾に書く。
インストール直後はアンインストールしたくなることが最も多い。

### アナリティクス {#analytics}

利用状況メトリクスは、ユーザーがどう使っているか、改善点、注力すべき領域の理解に役立つ。
しかし Web と違い、コマンドラインのユーザーは環境の制御を期待しており、裏で勝手に通信されると驚く。

**同意なしに使用状況やクラッシュデータを送信しない。**
ユーザーは必ず気づき、怒る。
収集する内容、理由、匿名化の方法、保持期間を明示する。

理想は「オプトイン」。
デフォルトで収集（オプトアウト）にするなら、Web サイトや初回起動で明確に伝え、無効化を簡単にする。

利用統計を収集するプロジェクト例:

- Angular.js は機能優先順位付けのために Google Analytics で詳細分析を収集。
  明示的なオプトインが必要。
  組織内で使う場合は追跡 ID を自分の Google Analytics に変えられる。
- Homebrew は Google Analytics にメトリクスを送信し、実践を詳述した [FAQ](https://docs.brew.sh/Analytics) がある。
- Next.js は匿名化された利用統計を収集し、デフォルトで有効。

**アナリティクス収集の代替を検討する。**

- Web ドキュメントを計測する。
  CLI ツールの使われ方を知りたいなら、知りたいユースケース中心のドキュメントを作り、その推移を見る。
  ドキュメント内の検索語を調べる。
- ダウンロードを計測する。
  利用量や OS の把握の荒い指標になる。
- ユーザーと話す。
  どう使っているか尋ねる。
  フィードバックや機能要望を docs や repo で促し、提出者から文脈を引き出す。

_参考: [Open Source Metrics](https://opensource.guide/metrics/)_

## 参考文献

- [The Unix Programming Environment](https://en.wikipedia.org/wiki/The_Unix_Programming_Environment), Brian W. Kernighan and Rob Pike
- [POSIX Utility Conventions](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html)
- [Program Behavior for All Programs](https://www.gnu.org/prep/standards/html_node/Program-Behavior.html), GNU Coding Standards
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46), Jeff Dickey
- [CLI Style Guide](https://devcenter.heroku.com/articles/cli-style-guide), Heroku

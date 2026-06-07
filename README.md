# aivoice2-srt

A.I.VOICE2 で出力した音声ファイル（WAV）とテキストファイル（TXT）から、SRT 字幕ファイルを自動生成する CLI ツールです。

## 機能

- 指定ディレクトリ内の同名 WAV/TXT ファイルのペアを自動検出
- WAV ファイルの再生時間を解析
- テキストを文単位（。！？!?.）で分割し、文字数に比例した時間配分で SRT を生成

## インストール

```bash
go install github.com/a-kaibu/aivoice2-srt@latest
```

## 使い方

```bash
aivoice2-srt <ディレクトリパス>
```

指定したディレクトリ内に、同じベース名を持つ `.wav` と `.txt` のペアがあれば、対応する `.srt` ファイルを同じディレクトリに出力します。

### 例

```
input/
├── narration01.wav
├── narration01.txt
├── narration02.wav
└── narration02.txt
```

```bash
aivoice2-srt ./input
```

実行後:

```
input/
├── narration01.wav
├── narration01.txt
├── narration01.srt  ← 生成
├── narration02.wav
├── narration02.txt
└── narration02.srt  ← 生成
```

## SRT 生成ロジック

テキストを句点・感嘆符・疑問符で文に分割し、各文の文字数に比例して音声全体の再生時間を割り当てます。

例: 再生時間 10 秒、「あいう。」（3文字）+「かきくけこ。」（5文字）の場合

- 「あいう。」→ 0:00 〜 3.75 秒
- 「かきくけこ。」→ 3.75 〜 10.00 秒

## ライセンス

MIT

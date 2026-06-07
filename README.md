# aivoice2-srt

A.I.VOICE2 で出力した音声ファイル（WAV）とテキストファイル（TXT）から、SRT 字幕ファイルを自動生成する CLI ツールです。

## 機能

- 指定ディレクトリ内の同名 WAV/TXT ファイルのペアを自動検出
- WAV ファイルの再生時間を解析
- 1ファイル = 1字幕エントリとして SRT を生成

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

A.I.VOICE2 はファイル単位で分割出力するため、1つの WAV/TXT ペアがそのまま1つの字幕エントリになります。WAV の再生時間が字幕の表示時間として使われます。

## ライセンス

MIT

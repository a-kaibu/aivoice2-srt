# aivoice2-srt

A.I.VOICE2 で出力した音声ファイル（WAV）とテキストファイル（TXT）から、SRT 字幕ファイルを自動生成する CLI ツールです。

## 機能

- 指定ディレクトリ内の同名 WAV/TXT ファイルのペアを自動検出
- WAV ファイルの再生時間を解析
- ファイル名順にソートし、時間を積み上げて1つの SRT ファイルに結合出力

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
├── narration02.wav
├── narration02.txt
└── input.srt  ← 生成（ディレクトリ名.srt）
```

## SRT 生成ロジック

ファイル名順にソートし、各 WAV の再生時間を積み上げて連番の字幕エントリを生成します。出力ファイル名はディレクトリ名 + `.srt` です。

## ライセンス

MIT

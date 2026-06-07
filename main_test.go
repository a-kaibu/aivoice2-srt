package main

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		seconds float64
		want    string
	}{
		{0, "00:00:00,000"},
		{1.5, "00:00:01,500"},
		{61.123, "00:01:01,123"},
		{3661.999, "01:01:01,999"},
		{7200, "02:00:00,000"},
	}

	for _, tt := range tests {
		got := formatTime(tt.seconds)
		if got != tt.want {
			t.Errorf("formatTime(%v) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestSplitSentences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "japanese periods",
			input: "こんにちは。世界。",
			want:  []string{"こんにちは。", "世界。"},
		},
		{
			name:  "mixed delimiters",
			input: "走れ！なぜ？終わり。",
			want:  []string{"走れ！", "なぜ？", "終わり。"},
		},
		{
			name:  "no delimiter at end",
			input: "始まり。途中",
			want:  []string{"始まり。", "途中"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   \n  ",
			want:  nil,
		},
		{
			name:  "english periods",
			input: "Hello. World.",
			want:  []string{"Hello.", "World."},
		},
		{
			name:  "single sentence no delimiter",
			input: "テスト",
			want:  []string{"テスト"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitSentences(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitSentences(%q) returned %d sentences, want %d\ngot: %v", tt.input, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("sentence[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGenerateSRT(t *testing.T) {
	sentences := []string{"あいう。", "かきくけこ。"}
	// 4 chars + 6 chars = 10 chars total, duration = 10s
	// first: 10 * 4/10 = 4s, second: 10 * 6/10 = 6s
	result := generateSRT(sentences, 10.0)

	expected := "1\n00:00:00,000 --> 00:00:04,000\nあいう。\n\n2\n00:00:04,000 --> 00:00:10,000\nかきくけこ。\n\n"
	if result != expected {
		t.Errorf("generateSRT mismatch\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

func TestGenerateSRT_SingleSentence(t *testing.T) {
	sentences := []string{"テスト。"}
	result := generateSRT(sentences, 5.0)

	expected := "1\n00:00:00,000 --> 00:00:05,000\nテスト。\n\n"
	if result != expected {
		t.Errorf("generateSRT mismatch\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

func TestWavDuration(t *testing.T) {
	dir := t.TempDir()
	wavPath := filepath.Join(dir, "test.wav")

	createTestWav(t, wavPath, 44100, 2, 16, 44100*2*2) // 1 second of audio

	duration, err := wavDuration(wavPath)
	if err != nil {
		t.Fatalf("wavDuration failed: %v", err)
	}

	if math.Abs(duration-1.0) > 0.01 {
		t.Errorf("wavDuration = %v, want ~1.0", duration)
	}
}

func TestWavDuration_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.wav")
	os.WriteFile(path, []byte("not a wav file at all"), 0644)

	_, err := wavDuration(path)
	if err == nil {
		t.Error("expected error for invalid WAV, got nil")
	}
}

func createTestWav(t *testing.T, path string, sampleRate uint32, channels uint16, bitsPerSample uint16, dataSize uint32) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	byteRate := sampleRate * uint32(channels) * uint32(bitsPerSample) / 8
	blockAlign := channels * bitsPerSample / 8
	fmtSize := uint32(16)
	fileSize := 4 + (8 + fmtSize) + (8 + dataSize)

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, fileSize)
	f.Write([]byte("WAVE"))

	// fmt chunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, fmtSize)
	binary.Write(f, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(f, binary.LittleEndian, channels)
	binary.Write(f, binary.LittleEndian, sampleRate)
	binary.Write(f, binary.LittleEndian, byteRate)
	binary.Write(f, binary.LittleEndian, blockAlign)
	binary.Write(f, binary.LittleEndian, bitsPerSample)

	// data chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, dataSize)
	// Write zeros for audio data
	zeros := make([]byte, dataSize)
	f.Write(zeros)
}

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


func TestGenerateSRT(t *testing.T) {
	result := generateSRT("こんにちは。", 10.0)

	expected := "1\n00:00:00,000 --> 00:00:10,000\nこんにちは。\n\n"
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

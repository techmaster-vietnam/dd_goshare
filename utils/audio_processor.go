package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertWavToOgg(inputPath, outputPath string) error {
	// Kiểm tra file input có tồn tại không
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("file input không tồn tại: %s", inputPath)
	}

	// Kiểm tra file có kích thước > 0 không
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("lỗi khi kiểm tra file: %s", inputPath)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("file WAV rỗng: %s", inputPath)
	}

	// Kiểm tra extension của file input
	if !strings.HasSuffix(inputPath, ".wav") {
		return fmt.Errorf("file input phải có extension .wav: %s", inputPath)
	}

	// Nếu outputPath trống, tự động tạo tên file output
	if outputPath == "" {
		outputPath = strings.TrimSuffix(inputPath, ".wav") + ".ogg"
	}

	// Kiểm tra ffmpeg có sẵn không
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg không được tìm thấy trong PATH: %v", err)
	}

	// Kiểm tra file WAV có hợp lệ không
	checkCmd := exec.Command("ffmpeg", "-v", "error", "-i", inputPath, "-f", "null", "-")
	checkOutput, checkErr := checkCmd.CombinedOutput()
	if checkErr != nil {
		return fmt.Errorf("file WAV không hợp lệ: %s, lỗi: %v, output: %s", inputPath, checkErr, string(checkOutput))
	}

	// Tạo thư mục output nếu chưa tồn tại
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục output: %v", err)
	}
	// Chạy lệnh ffmpeg để chuyển đổi với nén tối ưu cho dung lượng thấp nhất
	cmd := exec.Command("ffmpeg",
		"-i", inputPath, // Input file
		"-c:a", "libvorbis", // Audio codec
		"-q:a", "0", // Quality setting (0-10, 0 = dung lượng nhỏ nhất)
		"-b:a", "20k", // Bitrate tối đa 20kbps (tối thiểu cho voice)
		"-ar", "16000", // Sample rate 16kHz (tối thiểu cho voice)
		"-ac", "1", // Mono channel (giảm dung lượng)
		"-compression_level", "10", // Mức nén cao nhất cho Vorbis
		"-cutoff", "8000", // Cắt tần số trên 8kHz
		"-application", "voip", // Tối ưu cho voice
		"-y",       // Overwrite output file if exists
		outputPath, // Output file
	)

	// Chạy lệnh và lấy output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lỗi khi chạy ffmpeg: %v, output: %s", err, string(output))
	}

	return nil
}

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConvertWebP chuyển đổi ảnh sang định dạng WebP
// Nếu targetWidth = 640, sẽ tự động set tỷ lệ 16:9 (640x360px)
func ConvertWebP(inputImage, outputImage, preset string, quality, targetWidth int) error {
	// Xử lý preset mặc định
	if preset == "" {
		preset = "drawing"
	}

	// Kiểm tra preset hợp lệ
	validPresets := []string{"photo", "drawing", "icon"}
	isValidPreset := false
	for _, p := range validPresets {
		if p == preset {
			isValidPreset = true
			break
		}
	}
	if !isValidPreset {
		return fmt.Errorf("preset không hợp lệ. Phải là một trong: %v", validPresets)
	}

	// Xử lý quality mặc định
	if quality == 0 {
		quality = 80
	}

	// Kiểm tra quality trong khoảng hợp lệ
	if quality < 0 || quality > 100 {
		return fmt.Errorf("quality phải trong khoảng 0-100")
	}

	// Xử lý outputImage
	outputPath := outputImage

	// Kiểm tra xem outputImage có phải là đường dẫn file đầy đủ không
	if strings.Contains(outputImage, ".") {
		// Có file extension
		ext := strings.ToLower(filepath.Ext(outputImage))
		if ext != ".webp" {
			return fmt.Errorf("outputImage phải có extension .webp, nhận được: %s", ext)
		}
	} else {
		// Không có file extension, coi như là thư mục
		// Lấy tên file từ inputImage
		fileName := filepath.Base(inputImage)
		// Bỏ extension cũ và thêm .webp
		fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		outputPath = filepath.Join(outputImage, fileNameWithoutExt+".webp")
	}

	// Tạo thư mục output nếu chưa tồn tại
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục output: %v", err)
	}

	// Xây dựng lệnh cwebp
	var cmd *exec.Cmd
	if targetWidth > 0 {
		// Nếu targetWidth = 640, sử dụng tỷ lệ 16:9 (640x360px)
		if targetWidth == 640 {
			cmd = exec.Command("cwebp",
				"-preset", preset,
				"-q", fmt.Sprintf("%d", quality),
				"-resize", "640", "360",
				inputImage, "-o", outputPath)
		} else {
			cmd = exec.Command("cwebp",
				"-preset", preset,
				"-q", fmt.Sprintf("%d", quality),
				"-resize", fmt.Sprintf("%d", targetWidth), "0",
				inputImage, "-o", outputPath)
		}
	} else {
		cmd = exec.Command("cwebp",
			"-preset", preset,
			"-q", fmt.Sprintf("%d", quality),
			inputImage, "-o", outputPath)
	}

	// Thực thi lệnh
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lỗi khi chuyển đổi WebP: %v, output: %s", err, string(output))
	}

	return nil
}

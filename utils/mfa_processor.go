package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// RunMFAAlignment chạy lệnh MFA align từ thư mục data
func RunMFAAlignment() error {
	// Xác định thư mục data tuyệt đối
	dataDir := "../data"
	dictPath := filepath.Join(dataDir, "english_us_mfa.dict")
	if _, err := os.Stat(dictPath); os.IsNotExist(err) {
		return fmt.Errorf("dictionary file not found at %s", dictPath)
	}

	cmd := exec.Command("mfa", "align", "--clean", "--use_mp", "--output_format", "json",
		"corpus",
		"english_us_mfa.dict",
		"english_mfa.zip",
		"output")
	cmd.Dir = dataDir

	log.Println("Running MFA alignment from data directory...")

	// Kiểm tra xem MFA có sẵn không
	if _, err := exec.LookPath("mfa"); err != nil {
		return fmt.Errorf("MFA not found in PATH: %v", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("MFA command output: %s", string(output))
		return fmt.Errorf("MFA alignment failed: %v, output: %s", err, string(output))
	}

	log.Printf("MFA alignment completed successfully")
	return nil
}

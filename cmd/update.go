package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stacksenv/cli/version"
)

const (
	githubAPIURL = "https://api.github.com/repos/stacksenv/cli/releases/latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateCheckCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the stacksenv CLI",
	Long:  `Update the stacksenv CLI to the latest version.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return performUpdate()
	},
}

var updateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for updates",
	Long:  `Check if a newer version of stacksenv is available.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return checkForUpdates()
	},
}

// checkForUpdates checks if a newer version is available and displays the result.
func checkForUpdates() error {
	currentVersion := version.Version
	if currentVersion == "(untracked)" {
		fmt.Println("Current version: (development build)")
	} else {
		fmt.Printf("Current version: %s\n", currentVersion)
	}

	latestRelease, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")
	fmt.Printf("Latest version: %s\n", latestVersion)

	if currentVersion == "(untracked)" {
		fmt.Println("\nNote: You are running a development build. Update check may not be accurate.")
		return nil
	}

	if compareVersions(currentVersion, latestVersion) < 0 {
		fmt.Printf("\n✓ Update available! Run 'stacksenv update' to update to version %s\n", latestVersion)
	} else {
		fmt.Println("\n✓ You are running the latest version")
	}

	return nil
}

// performUpdate downloads and installs the latest version of stacksenv.
func performUpdate() error {
	currentVersion := version.Version
	fmt.Printf("Current version: %s\n", currentVersion)

	latestRelease, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")
	fmt.Printf("Latest version: %s\n", latestVersion)

	if currentVersion != "(untracked)" && compareVersions(currentVersion, latestVersion) >= 0 {
		fmt.Println("You are already running the latest version")
		return nil
	}

	// Determine OS and architecture
	osName, arch := getOSArch()
	fmt.Printf("Detected platform: %s/%s\n", osName, arch)

	// Find the appropriate asset
	assetURL, assetName, err := findAsset(latestRelease, osName, arch)
	if err != nil {
		return fmt.Errorf("failed to find release asset: %w", err)
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Download the archive
	tmpDir, err := os.MkdirTemp("", "stacksenv-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(assetURL, archivePath); err != nil {
		return fmt.Errorf("failed to download release: %w", err)
	}

	fmt.Println("Extracting...")

	// Extract the binary
	binaryName := "stacksenv"
	if osName == "windows" {
		binaryName = "stacksenv.exe"
	}
	binaryPath := filepath.Join(tmpDir, binaryName)

	if strings.HasSuffix(assetName, ".zip") {
		if err := extractZip(archivePath, binaryPath, binaryName); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	} else {
		if err := extractTarGz(archivePath, binaryPath, binaryName); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Make binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	fmt.Printf("Installing to %s...\n", execPath)

	// Replace the current binary
	if err := replaceBinary(binaryPath, execPath); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	fmt.Printf("Successfully updated to version %s\n", latestVersion)
	return nil
}

// getLatestRelease fetches the latest release information from GitHub API.
func getLatestRelease() (*githubRelease, error) {
	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// getOSArch returns the OS and architecture names matching the release asset naming.
func getOSArch() (string, string) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go OS names to release asset names
	osMap := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux",
		"windows": "windows",
		"freebsd": "freebsd",
	}

	osName := osMap[goos]
	if osName == "" {
		osName = goos
	}

	// Map Go arch names to release asset names
	archMap := map[string]string{
		"amd64": "amd64",
		"386":   "386",
		"arm64": "arm64",
		"arm":   "armv7", // Default to armv7, could be improved
	}

	arch := archMap[goarch]
	if arch == "" {
		arch = goarch
	}

	return osName, arch
}

// findAsset finds the appropriate release asset for the given OS and architecture.
func findAsset(release *githubRelease, osName, arch string) (string, string, error) {
	expectedName := fmt.Sprintf("%s-%s-stacksenv", osName, arch)
	if osName == "windows" {
		expectedName += ".zip"
	} else {
		expectedName += ".tar.gz"
	}

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset.BrowserDownloadURL, asset.Name, nil
		}
	}

	return "", "", fmt.Errorf("no asset found for %s/%s", osName, arch)
}

// downloadFile downloads a file from a URL to a local path.
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarGz extracts a binary from a tar.gz archive.
func extractTarGz(archivePath, destPath, binaryName string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == binaryName {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			if _, err := io.Copy(out, tr); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}

// extractZip extracts a binary from a zip archive.
func extractZip(archivePath, destPath, binaryName string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == binaryName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			return err
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}

// replaceBinary replaces the current executable with the new binary.
func replaceBinary(newBinary, currentExec string) error {
	// On Windows, we need to remove the old file first
	if runtime.GOOS == "windows" {
		if err := os.Remove(currentExec); err != nil && !os.IsNotExist(err) {
			return err
		}
		return os.Rename(newBinary, currentExec)
	}

	// On Unix-like systems, we can use rename which is atomic
	return os.Rename(newBinary, currentExec)
}

// compareVersions compares two version strings.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Simple string comparison for semantic versions
	// This works for versions like "1.0.0", "1.0.1", etc.
	if v1 == v2 {
		return 0
	}
	if v1 < v2 {
		return -1
	}
	return 1
}

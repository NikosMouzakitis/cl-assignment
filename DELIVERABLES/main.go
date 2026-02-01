package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const imageTag = "stream8-kernel-builder:local"

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: build-stream8-kernel <src.rpm|url> <outfolder>\n")
	os.Exit(2)
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ensureDocker() error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return errors.New("docker not found in PATH")
	}
	return nil
}

func abs(p string) string {
	ap, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return ap
}

func downloadToTemp(u string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "srpm-*")
	if err != nil {
		return "", err
	}
	dst := filepath.Join(tmpDir, "input.src.rpm")
	if err := run("curl", "-fL", "-o", dst, u); err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	return dst, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func main() {
	if len(os.Args) != 3 {
		usage()
	}
	src := os.Args[1]
	outDir := abs(os.Args[2])

	if err := ensureDocker(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating output dir:", err)
		os.Exit(1)
	}

	var srpmPath string
	if isURL(src) {
		fmt.Println("[*] SRPM is a URL, downloading...")
		p, err := downloadToTemp(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		srpmPath = p
	} else {
		srpmPath = abs(src)
		if _, err := os.Stat(srpmPath); err != nil {
			fmt.Fprintln(os.Stderr, "Error: SRPM not found:", srpmPath)
			os.Exit(1)
		}
	}

	if sum, err := sha256File(srpmPath); err == nil {
		fmt.Println("[*] SRPM sha256:", sum)
	}

	// Build Docker image from current directory (expects Dockerfile here)
	fmt.Println("[*] Building Docker image:", imageTag)
	if err := run("docker", "build", "-t", imageTag, "."); err != nil {
		fmt.Fprintln(os.Stderr, "Docker build failed:", err)
		os.Exit(1)
	}

	// Mount SRPM directory read-only, mount output writable.
	srpmDir := filepath.Dir(srpmPath)
	srpmBase := filepath.Base(srpmPath)

	uid := os.Getuid()
	gid := os.Getgid()

	fmt.Println("[*] Running build in container...")
	args := []string{
		"run", "--rm",
		"-u", "0", // root (dnf builddep needs it)
		"-e", fmt.Sprintf("HOST_UID=%d", uid),
		"-e", fmt.Sprintf("HOST_GID=%d", gid),
		"-v", srpmDir + ":/in:ro",
		"-v", outDir + ":/out",
		imageTag,
		"/bin/bash", "-lc",
		strings.Join([]string{
			"set -euo pipefail",
			"rm -rf /root/rpmbuild && rpmdev-setuptree",
			"rpm -ivh /in/" + srpmBase,
			"spec=$(ls -1 /root/rpmbuild/SPECS/*.spec | head -n1)",
			"dnf -y builddep \"$spec\"",
			// rebuild SRPM => produces RPMS automatically
			"rpmbuild --rebuild /in/" + srpmBase,
			"mkdir -p /out",
			"shopt -s globstar nullglob",
			"cp -av /root/rpmbuild/RPMS/**/*.rpm /out/ || true",
			"cp -av /root/rpmbuild/SRPMS/*.src.rpm /out/ || true",
			"chown -R " + fmt.Sprintf("%d:%d", uid, gid) + " /out || true",
			"echo '[*] Exported artifacts:'",
			"ls -lh /out || true",
		}, "\n"),
	}

	if err := run("docker", args...); err != nil {
		fmt.Fprintln(os.Stderr, "Build failed:", err)
		os.Exit(1)
	}

	fmt.Println("[*] Done. Output in:", outDir)
}


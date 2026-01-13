package skills

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// PackageSkillDir zips a skill directory into a .skill file.
func PackageSkillDir(skillDir string, outDir string, force bool) (string, error) {
	if strings.TrimSpace(skillDir) == "" {
		return "", errors.New("skill directory is required")
	}
	if strings.TrimSpace(outDir) == "" {
		return "", errors.New("output directory is required")
	}

	info, err := os.Stat(skillDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to stat skill directory")
	}
	if !info.IsDir() {
		return "", errors.New("skill path must be a directory")
	}

	skillName := filepath.Base(skillDir)
	outPath := filepath.Join(outDir, skillName+".skill")
	if _, err := os.Stat(outPath); err == nil && !force {
		return "", errors.Errorf("output file already exists: %s", outPath)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", errors.Wrap(err, "failed to create output directory")
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to create output file")
	}
	defer func() {
		_ = outFile.Close()
	}()

	zipWriter := zip.NewWriter(outFile)
	defer func() {
		_ = zipWriter.Close()
	}()

	parentDir := filepath.Dir(skillDir)

	err = filepath.WalkDir(skillDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		fileInfo, err := d.Info()
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(parentDir, path)
		if err != nil {
			return err
		}
		arcPath := filepath.ToSlash(relPath)

		head, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		head.Name = arcPath
		head.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(head)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		if _, err := io.Copy(writer, file); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to package skill")
	}

	return outPath, nil
}

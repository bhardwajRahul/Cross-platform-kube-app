package mage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/magefile/mage/sh"
)

type rpmTemplateData struct {
	AppShortName string
	AppLongName  string
	Version      string
	RPMVersion   string
	RPMArch      string
}

type rpmPackageContext struct {
	templateData   rpmTemplateData
	linuxOutputDir string
	topDir         string
	rpmArch        string
	rpmVersion     string
}

func renderRPMTemplate(templatePath, destPath string, data rpmTemplateData, perm os.FileMode) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", destPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to render %s: %w", destPath, err)
	}

	if perm != 0 {
		if err := os.Chmod(destPath, perm); err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", destPath, err)
		}
	}

	return nil
}

func packageRPM(cfg BuildConfig, binPath, packagesDir string) error {
	if _, err := exec.LookPath("rpmbuild"); err != nil {
		return fmt.Errorf("rpmbuild not found; install rpm-build to build .rpm packages")
	}

	fmt.Println("\n📦 Building .rpm package")
	packaging, err := prepareRPMPackage(cfg, packagesDir)
	if err != nil {
		return err
	}
	if err := stageRPMPackage(cfg, binPath, packaging); err != nil {
		return err
	}
	specPath := filepath.Join(packaging.topDir, "SPECS", fmt.Sprintf("%s.spec", cfg.AppShortName))
	if err := sh.RunV("rpmbuild", "-bb", "--quiet", "--define", fmt.Sprintf("_topdir %s", packaging.topDir), specPath); err != nil {
		return fmt.Errorf("failed to build .rpm: %w", err)
	}
	return collectRPMPackage(cfg, packaging)
}

func prepareRPMPackage(cfg BuildConfig, packagesDir string) (rpmPackageContext, error) {
	rpmArch := map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}[cfg.ArchType]
	if rpmArch == "" {
		rpmArch = cfg.ArchType
	}
	rpmVersion := strings.ReplaceAll(cfg.Version, "-", ".")
	templateData := rpmTemplateData{
		AppShortName: cfg.AppShortName,
		AppLongName:  cfg.AppLongName,
		Version:      cfg.Version,
		RPMVersion:   rpmVersion,
		RPMArch:      rpmArch,
	}

	linuxOutputDir := filepath.Join(cfg.BuildDir, "artifacts")
	if err := os.MkdirAll(linuxOutputDir, 0o755); err != nil {
		return rpmPackageContext{}, fmt.Errorf("failed to create linux package output dir: %w", err)
	}

	topDir := filepath.Join(packagesDir, "rpm")
	topDirAbs, err := filepath.Abs(topDir)
	if err != nil {
		return rpmPackageContext{}, fmt.Errorf("failed to resolve rpm topdir: %w", err)
	}
	dirs := []string{
		filepath.Join(topDirAbs, "BUILD"),
		filepath.Join(topDirAbs, "RPMS"),
		filepath.Join(topDirAbs, "SOURCES"),
		filepath.Join(topDirAbs, "SPECS"),
		filepath.Join(topDirAbs, "SRPMS"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return rpmPackageContext{}, fmt.Errorf("failed to create rpm dir %s: %w", dir, err)
		}
	}
	return rpmPackageContext{
		templateData:   templateData,
		linuxOutputDir: linuxOutputDir,
		topDir:         topDirAbs,
		rpmArch:        rpmArch,
		rpmVersion:     rpmVersion,
	}, nil
}

func stageRPMPackage(cfg BuildConfig, binPath string, packaging rpmPackageContext) error {
	sourceBin := filepath.Join(packaging.topDir, "SOURCES", cfg.AppShortName)
	if err := sh.Copy(sourceBin, binPath); err != nil {
		return fmt.Errorf("failed to copy binary into rpm SOURCES: %w", err)
	}

	sourceDesktop := filepath.Join(packaging.topDir, "SOURCES", fmt.Sprintf("%s.desktop", cfg.AppShortName))
	desktopTemplate := filepath.Join("mage", "rpm", "desktop.tmpl")
	if err := renderRPMTemplate(desktopTemplate, sourceDesktop, packaging.templateData, 0o644); err != nil {
		return err
	}

	if _, err := os.Stat(cfg.IconSource); err != nil {
		return fmt.Errorf("icon not found at %s: %w", cfg.IconSource, err)
	}
	sourceIcon := filepath.Join(packaging.topDir, "SOURCES", fmt.Sprintf("%s.png", cfg.AppShortName))
	if err := sh.Copy(sourceIcon, cfg.IconSource); err != nil {
		return fmt.Errorf("failed to copy icon into rpm SOURCES: %w", err)
	}

	sourceMetainfo := filepath.Join(packaging.topDir, "SOURCES", fmt.Sprintf("%s.metainfo.xml", cfg.AppShortName))
	metainfoTemplate := filepath.Join("mage", "rpm", "metainfo.xml.tmpl")
	if err := renderRPMTemplate(metainfoTemplate, sourceMetainfo, packaging.templateData, 0o644); err != nil {
		return err
	}

	specPath := filepath.Join(packaging.topDir, "SPECS", fmt.Sprintf("%s.spec", cfg.AppShortName))
	specTemplate := filepath.Join("mage", "rpm", "spec.tmpl")
	if err := renderRPMTemplate(specTemplate, specPath, packaging.templateData, 0o644); err != nil {
		return err
	}
	return nil
}

func collectRPMPackage(cfg BuildConfig, packaging rpmPackageContext) error {
	rpmOutputDir := filepath.Join(packaging.topDir, "RPMS", packaging.rpmArch)
	matches, _ := filepath.Glob(filepath.Join(rpmOutputDir, fmt.Sprintf("%s-%s-1.*.rpm", cfg.AppShortName, packaging.rpmVersion)))
	if len(matches) > 0 {
		rpmFile := matches[0]
		rpmCopy := filepath.Join(packaging.linuxOutputDir, fmt.Sprintf("%s-%s-linux-%s.rpm", cfg.AppShortName, cfg.Version, packaging.rpmArch))
		if err := sh.Copy(rpmCopy, rpmFile); err != nil {
			return fmt.Errorf("failed to copy rpm into %s: %w", rpmCopy, err)
		}
		if err := os.RemoveAll(packaging.topDir); err != nil {
			return fmt.Errorf("failed to clean rpm staging dir: %w", err)
		}
		fmt.Printf("✅ Built RPM package: %s\n", rpmCopy)
	} else {
		_ = os.RemoveAll(packaging.topDir)
		fmt.Printf("✅ Built RPM package under %s\n", rpmOutputDir)
	}
	return nil
}

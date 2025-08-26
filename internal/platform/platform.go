package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/clix-so/nativefire/internal/firebase"
)

type Type int

const (
	Android Type = iota
	iOS
	MacOS
	Windows
	Linux
)

// Platform name constants
const (
	iosString = "ios"
)

type Platform interface {
	Name() string
	Type() Type
	Detect() bool
	InstallConfig(config *firebase.Config) error
	AddInitializationCode(config *firebase.Config) error
	ConfigFileName() string
	ConfigPath() string
}

type AndroidPlatform struct{}
type IOSPlatform struct{}
type MacOSPlatform struct{}
type WindowsPlatform struct{}
type LinuxPlatform struct{}

func DetectPlatform() (Platform, error) {
	platforms := []Platform{
		&AndroidPlatform{},
		&IOSPlatform{},
		&MacOSPlatform{},
		&WindowsPlatform{},
		&LinuxPlatform{},
	}

	for _, platform := range platforms {
		if platform.Detect() {
			return platform, nil
		}
	}

	return nil, fmt.Errorf("no supported platform detected in current directory")
}

func FromString(platformStr string) (Platform, error) {
	switch strings.ToLower(platformStr) {
	case "android":
		return &AndroidPlatform{}, nil
	case iosString:
		return &IOSPlatform{}, nil
	case "macos":
		return &MacOSPlatform{}, nil
	case "windows":
		return &WindowsPlatform{}, nil
	case "linux":
		return &LinuxPlatform{}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platformStr)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func findFile(root, pattern string) string {
	var result string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return nil
		}
		if matched {
			result = path
			return filepath.SkipDir
		}
		return nil
	})
	return result
}

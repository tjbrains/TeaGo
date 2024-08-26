package Tea

import (
	"github.com/tjbrains/TeaGo/logs"
	"github.com/tjbrains/TeaGo/utils/string"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

var Env = EnvDev
var DS = string(os.PathSeparator)

var publicDir string
var viewsDir string
var configDir string
var tmpDir string
var isTesting = checkIsTesting()

var Root string

// 初始化
func init() {
	Root = findRoot()
}

// 判断是否在某个特定环境下
func Is(env ...string) bool {
	if len(env) == 0 {
		return false
	}
	for _, envItem := range env {
		if envItem == Env {
			return true
		}
	}
	return false
}

// 判断是否在测试模式下
func IsTesting() bool {
	return isTesting
}

func checkIsTesting() bool {
	execFile := filepath.Base(os.Args[0])
	return execFile == "main" ||
		execFile == "main.exe" ||
		strings.HasPrefix(execFile, "___") ||
		strings.HasSuffix(execFile, ".test") ||
		strings.HasSuffix(execFile, ".test.exe")
}

// 取得临时目录
func TmpDir() string {
	if len(tmpDir) > 0 {
		return tmpDir
	}
	return Root + DS + "tmp"
}

// 设置临时目录
func SetTmpDir(dir string) {
	tmpDir = dir
}

// 取得临时文件
func TmpFile(file string) string {
	if runtime.GOOS == "windows" {
		file = strings.Replace(file, "/", DS, -1)
	}
	return TmpDir() + DS + file
}

func LogDir() string {
	return Root + DS + "logs"
}

func LogFile(file string) string {
	if runtime.GOOS == "windows" {
		file = strings.Replace(file, "/", DS, -1)
	}
	return LogDir() + DS + file
}

func BinDir() string {
	return Root + DS + "bin"
}

func PublicDir() string {
	if len(publicDir) > 0 {
		return publicDir
	}

	publicDir = findLatestDir(Root, "public")
	return publicDir
}

func SetPublicDir(dir string) {
	publicDir = dir
}

func PublicFile(file string) string {
	if runtime.GOOS == "windows" {
		file = strings.Replace(file, "/", DS, -1)
	}
	return PublicDir() + DS + file
}

func ViewsDir() string {
	if len(viewsDir) > 0 {
		return viewsDir
	}

	viewsDir = findLatestDir(Root, "views")
	return viewsDir
}

func SetViewsDir(dir string) {
	viewsDir = dir
}

func ConfigDir() string {
	if len(configDir) > 0 {
		return configDir
	}

	configDir = findLatestDir(Root, "configs")
	return configDir
}

func SetConfigDir(dir string) {
	configDir = dir
}

func ConfigFile(file string) string {
	if runtime.GOOS == "windows" {
		file = strings.Replace(file, "/", DS, -1)
	}
	return ConfigDir() + DS + file
}

func findRoot() string {
	// TEAROOT变量
	root := strings.TrimSpace(os.Getenv("TEAROOT"))
	if len(root) > 0 {
		abs, err := filepath.Abs(root)
		if err != nil {
			logs.Errorf("invalid GOPATH '%s'", root)
			return root
		}
		return abs
	}

	// GOPATH变量
	if IsTesting() {
		root = strings.TrimSpace(os.Getenv("GOPATH"))
		if len(root) > 0 {
			abs, err := filepath.Abs(root)
			if err != nil {
				logs.Errorf("invalid GOPATH '%s'", root)
				return root + DS + "src" + DS + "main"
			}
			return abs + DS + "src" + DS + "main"
		}
	}

	// 当前执行的目录
	dir, err := os.Getwd()
	if err == nil {
		return dir
	}
	return "./"
}

func UpdateRoot(root string) {
	Root = root
	publicDir = ""
	viewsDir = ""
	configDir = ""
}

func findLatestDir(parent string, name string) string {
	matches, err := filepath.Glob(parent + DS + name + ".*")
	if err != nil {
		logs.Errorf("%s", err.Error())
		return parent + DS + name
	}

	if len(matches) == 0 {
		return parent + DS + name
	}

	var lastVersion = ""
	var resultDir = ""

	for _, match := range matches {
		dirname := match
		stat, err := os.Stat(dirname)
		if err != nil || !stat.IsDir() {
			continue
		}

		version := filepath.Base(match)[len(name)+1:]

		if len(lastVersion) == 0 {
			lastVersion = version
			resultDir = dirname
			continue
		}

		if stringutil.VersionCompare(lastVersion, version) < 0 {
			lastVersion = version
			resultDir = dirname
			continue
		}
	}

	if len(resultDir) == 0 {
		return parent + DS + name
	}

	return resultDir
}

package version

import (
	"fmt"
	"runtime"
)

// 这些变量将在编译时通过 -ldflags 被覆盖
var (
	gitVersion   = "v0.0.0-dev"           // 默认版本号，用于开发阶段
	gitCommit    = "none"                 // Git 提交哈希
	gitTreeState = ""                     // Git 仓库状态，如 "clean" 或 "dirty"
	buildDate    = "1970-01-01T00:00:00Z" // 构建时间戳
)

// Info 结构体包含了完整的版本信息
type Info struct {
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// String 返回格式化的版本字符串
func (i Info) String() string {
	return fmt.Sprintf("version: %s\nbuildDate: %s\ngitCommit: %s\ngitTreeState: %s\ngoVersion: %s\ncompiler: %s\nplatform: %s",
		i.GitVersion,
		i.BuildDate,
		i.GitCommit,
		i.GitTreeState,
		i.GoVersion,
		i.Compiler,
		i.Platform,
	)
}

// Get 返回填充好的 Info 结构体
func Get() Info {
	return Info{
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

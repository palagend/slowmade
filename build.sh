#!/bin/bash

# 项目根目录
PROJECT_ROOT=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
# 输出的二进制文件名称
APP_NAME="slowmade"
# 输出的目录
OUTPUT_DIR="${PROJECT_ROOT}/.output"
# 版本包的导入路径，必须与你的代码中一致
VERSION_PACKAGE="github.com/palagend/slowmade/internal/version" # 请替换为你的模块名和版本包路径

# 获取版本信息
# 使用 git describe 获取最近的标签作为版本号，如果找不到则使用 commit hash
if [[ -z "${VERSION}" ]]; then
  VERSION=$(git describe --tags --always --match='v*' 2>/dev/null || echo "v0.0.0-unknown")
fi

# 检查 Git 工作树是否干净（是否有未提交的修改）
GIT_TREE_STATE="dirty"
if git status --porcelain 2>/dev/null | grep -q '^.*'; then
  GIT_TREE_STATE="dirty"
else
  GIT_TREE_STATE="clean"
fi

# 获取完整的 Git Commit Hash
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# 构建链接器标志 (ldflags)
# -X 标志用于在编译时设置指定变量的值
GO_LDFLAGS="-X ${VERSION_PACKAGE}.gitVersion=${VERSION} \
  -X ${VERSION_PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${VERSION_PACKAGE}.gitTreeState=${GIT_TREE_STATE} \
  -X ${VERSION_PACKAGE}.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"

echo "Building ${APP_NAME}..."
echo "Version: ${VERSION}"
echo "GitCommit: ${GIT_COMMIT}"
echo "GitTreeState: ${GIT_TREE_STATE}"
echo "=================================="
# 执行构建命令
cd ${PROJECT_ROOT}
go build -v -ldflags "${GO_LDFLAGS}" -o ${OUTPUT_DIR}/${APP_NAME} ./main.go
//go:build cgo && ((linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package database

// 通过空白导入注册 libSQL 驱动 (driver name "libsql")。
// 该驱动依赖预编译的原生库，仅在 linux/darwin (amd64/arm64) + CGO 环境下可用。
import _ "github.com/tursodatabase/go-libsql"

// libsqlAvailable 表示当前平台是否链接了 libSQL 驱动。
func libsqlAvailable() bool { return true }

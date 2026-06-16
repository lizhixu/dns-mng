//go:build !(cgo && ((linux && (amd64 || arm64)) || (darwin && (amd64 || arm64))))

package database

// libsqlAvailable 在不支持 libSQL 原生库的平台 (如 Windows) 上返回 false。
// 这些平台仍可使用默认的本地 SQLite (modernc.org/sqlite, 纯 Go)。
func libsqlAvailable() bool { return false }

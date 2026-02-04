package sqltils

import (
	"devops-console-backend/pkg/configs"
	"strings"

	"github.com/xwb1989/sqlparser"
)

// SplitAllStatements 分割sql语句到多个字符串
func SplitAllStatements(sql string) ([]string, error) {
	var stmts []string
	for {
		stmt, rest, err := sqlparser.SplitStatement(sql)
		if err != nil {
			return nil, err
		}
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			stmts = append(stmts, stmt)
		}
		if strings.TrimSpace(rest) == "" {
			break
		}
		sql = rest
	}

	return stmts, nil
}

// ExecSQLScript 执行sql语句
func ExecSQLScript(sqlContent string) error {
	stmts, err := SplitAllStatements(sqlContent)
	if err != nil {
		return err
	}

	for _, stmt := range stmts {
		configs.GORMDB.Exec(stmt)
	}
	return nil
}

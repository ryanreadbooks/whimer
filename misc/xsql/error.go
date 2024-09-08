package xsql

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNoRecord   = sqlx.ErrNotFound
	ErrDuplicate  = fmt.Errorf("duplicate entry")
	ErrOutOfRange = fmt.Errorf("out of range")
)

// 转换not found和duplicate entry两种错误
func ConvertError(err error) error {
	switch err {
	case nil:
		return nil
	case sqlx.ErrNotFound:
		return ErrNoRecord
	default:
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return ErrDuplicate
		}

		return err
	}
}

func IsMildErr(err error) bool {
	if err == nil {
		return true
	}
	return errors.Is(err, ErrNoRecord) || errors.Is(err, ErrDuplicate)
}

func IsCriticalErr(err error) bool {
	if err == nil {
		return false
	}

	return !IsMildErr(err)
}

func IsDuplicate(err error) bool {
	return errors.Is(err, ErrDuplicate)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNoRecord)
}

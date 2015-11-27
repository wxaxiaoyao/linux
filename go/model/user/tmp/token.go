package user

import (
	"fmt"

	"sirendaou.com/duserver/common"
	"sirendaou.com/duserver/common/errors"
)

func DBUniInsertToken(uid uint64, token string) error {
	sqlStr := fmt.Sprintf("insert into t_devicetoken (id, uid, token) values (0 , %d, '%s') on duplicate key update token = '%s'", uid, token, token)

	return errors.As(common.MysqlExec(sqlStr), uid, token)
}

func DBClearExtraToken(uid uint64, token string) error {
	sqlStr := ""

	if uid > 0 {
		sqlStr = fmt.Sprintf("update t_devicetoken set token = '' where token = '%s' and uid != %d ", token, uid)
	} else {
		sqlStr = fmt.Sprintf("update t_devicetoken set token = '' where token = '%s'", token)
	}

	return errors.As(common.MysqlExec(sqlStr), uid, token)
}

func DBGetUidByToken(token string) (uint64, error) {
	var uid uint64 = 0
	sqlStr := fmt.Sprintf("select uid from t_devicetoken where token = '%d'", token)

	rows, err := common.MysqlQuery(sqlStr)
	if err != nil {
		return 0, errors.As(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&uid); err != nil {
			return uid, errors.As(err, token)
		}
	}

	return uid, nil
}

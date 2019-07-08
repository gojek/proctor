package postgres

import "database/sql"

func StringToSQLString(str string) sql.NullString {
	if len(str) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: str,
		Valid:  true,
	}
}

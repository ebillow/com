package db

//
//type DBRow struct {
//	rets map[string][]byte
//	ok   bool
//}
//
//func fetchOneRow(rows *sql.Rows) (record map[string][]byte, ok bool) {
//	if rows.Next() {
//		columns, _ := rows.Columns()
//		scanArgs := make([]interface{}, len(columns))
//		values := make([]interface{}, len(columns))
//		for i := range values {
//			scanArgs[i] = &values[i]
//		}
//		ok = false
//		record = make(map[string][]byte, len(columns))
//
//		err := rows.Scan(scanArgs...)
//		if err != nil {
//			log.Warnf("fetch row error:%v", err)
//			return
//		}
//		for i, col := range values {
//			if col != nil {
//				record[columns[i]] = col.([]byte)
//			}
//		}
//		ok = true
//	}
//	return
//}
//
//func NewDBRow(rows *sql.Rows) (*DBRow, bool) {
//	r := &DBRow{}
//	r.rets, r.ok = fetchOneRow(rows)
//	return r, r.ok
//}
//
//func (r *DBRow) HasRet() bool {
//	return r.ok
//}
//
//func (r *DBRow) checkErr(col string) {
//	log.Errorf("get col error:%s", col)
//}
//
//func (r *DBRow) GetColInt(col string) int {
//	v, ok := r.rets[col]
//	if ok {
//		ret, err := strconv.Atoi(string(v))
//		ok = (err == nil)
//		return ret
//	}
//	r.checkErr(col)
//	return 0
//}
//func (r *DBRow) GetColInt64(col string) int64 {
//	v, ok := r.rets[col]
//	if ok {
//		ret, err := strconv.ParseInt(string(v), 10, 64)
//		ok = (err == nil)
//		return ret
//	}
//	r.checkErr(col)
//	return 0
//}
//
//func (r *DBRow) GetColBool(col string) bool {
//	return r.GetColInt(col) != 0
//}
//
//func (r *DBRow) GetColFloat32(col string) float32 {
//	v, ok := r.rets[col]
//	if ok {
//		ret, err := strconv.ParseFloat(string(v), 32)
//		ok = (err == nil)
//		return float32(ret)
//	}
//	r.checkErr(col)
//	return 0
//}
//func (r *DBRow) GetColFloat64(col string) float64 {
//	v, ok := r.rets[col]
//	if ok {
//		ret, err := strconv.ParseFloat(string(v), 64)
//		ok = (err == nil)
//		return ret
//	}
//	r.checkErr(col)
//	return 0
//}
//func (r *DBRow) GetColByte(col string) (v []byte) {
//	v, ok := r.rets[col]
//	if ok {
//		return v
//	}
//	r.checkErr(col)
//	return v
//}
//func (r *DBRow) GetColStr(col string) (s string) {
//	v, ok := r.rets[col]
//	if ok {
//		s = string(v)
//	}
//	return
//}

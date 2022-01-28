package structs

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

func StructifyMap(objMap *map[string]interface{}, obj interface{}) bool {
	fieldMap := GetFieldMap(obj)
	for key, value := range *objMap {
		switch value.(type) {
		case nil:
			continue

		case []byte:
			s := string(value.([]byte))
			x, err := strconv.Atoi(s)

			if err != nil {
				fieldMap[key].SetString(s)
			} else {
				fieldMap[key].SetInt((int64(x)))
			}

		case string:
			fieldMap[key].SetString(value.(string))

		case bool:
			fieldMap[key].SetBool(value.(bool))

		case float64:
			fieldMap[key].SetInt(int64(value.(float64)))

		default:
			return false
		}

	}
	return true
}

func StructifyRows(rows *sql.Rows, objType reflect.Type) (objs []interface{}, ok bool) {
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]interface{}, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	objs = make([]interface{}, 0)
	for rows.Next() {

		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		newObj := reflect.New(objType).Interface()
		fieldMap := GetFieldMap(newObj)
		objs = append(objs, newObj)

		for i, value := range values {
			switch value.(type) {
			case nil:
				continue

			case []byte:
				s := string(value.([]byte))
				x, err := strconv.Atoi(s)

				if err != nil {
					fieldMap[columns[i]].SetString(s)
				} else {
					fieldMap[columns[i]].SetInt((int64(x)))
				}

			case string:
				fieldMap[columns[i]].SetString(value.(string))

			case bool:
				fieldMap[columns[i]].SetBool(value.(bool))

			case int64:
				fieldMap[columns[i]].SetInt(value.(int64))

			default:
				ok = false
				return
			}
		}
	}
	ok = true
	return
}

func JsonifyRows(rows *sql.Rows) (jsonString string) {
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]interface{}, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	flag := false
	results := make(map[string]interface{})

	jsonString = "["
	for rows.Next() {
		if flag {
			//data = append(data, ",")
			jsonString += ","
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		for i, value := range values {
			switch value.(type) {
			case nil:
				results[columns[i]] = nil

			case []byte:
				s := string(value.([]byte))
				x, err := strconv.Atoi(s)

				if err != nil {
					results[columns[i]] = s
				} else {
					results[columns[i]] = x
				}

			default:
				results[columns[i]] = value
			}
		}

		b, _ := json.Marshal(results)
		//data = append(data, strings.TrimSpace(string(b)))
		jsonString += strings.TrimSpace(string(b))
		flag = true
	}
	jsonString += "]"

	return
}

func JsonifyRedisZ(set *[]redis.Z) (jsonString string) {
	jsonString = "{"

	jsonSenderIds := `"senderids":[`
	jsonSenderScores := `"senderscores":[`

	jsonGroupIds := `"groupids":[`
	jsonGroupScores := `"groupscores":[`

	for _, z := range *set {
		member := z.Member.(string)
		score := z.Score

		if member[0:2] == "DM" {
			jsonSenderIds += member[3:] + ","
			jsonSenderScores += strconv.FormatFloat(score, 'f', -1, 64) + ","
		} else if member[0:2] == "GM" {
			jsonGroupIds += member[3:] + ","
			jsonGroupScores += strconv.FormatFloat(score, 'f', -1, 64) + ","
		}
	}

	if jsonSenderIds[len(jsonSenderIds)-1] == ',' {
		jsonSenderIds = jsonSenderIds[:len(jsonSenderIds)-1] + "],"
		jsonSenderScores = jsonSenderScores[:len(jsonSenderScores)-1] + "],"
	} else {
		jsonSenderIds += "],"
		jsonSenderScores += "],"
	}

	if jsonGroupIds[len(jsonGroupIds)-1] == ',' {
		jsonGroupIds = jsonGroupIds[:len(jsonGroupIds)-1] + "],"
		jsonGroupScores = jsonGroupScores[:len(jsonGroupScores)-1] + "]"
	} else {
		jsonGroupIds += "],"
		jsonGroupScores += "]"
	}

	jsonString += jsonSenderIds + jsonSenderScores + jsonGroupIds + jsonGroupScores + "}"

	return
}

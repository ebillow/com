package mgo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"server/com/log"
)

func getIndexNames(collection *mongo.Collection) (names map[string]bool, err error) {
	ctx := context.Background()
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(context.Background(), &result); err != nil {
		return nil, err
	}

	names = make(map[string]bool)
	for i := 0; i != len(result); i++ {
		for k, v := range result[i] {
			if k == "name" {
				names[v.(string)] = true
			}
		}
	}
	return
}

// CreateIndex 创建索引，可以新增，无法修改已存在的索引。
//
//	参数createIDXs map的key为索引名，如：acc_1_worldid_1 表示索引key是acc,value=1和worldid,value=1
func CreateIndex(db *mongo.Database, table string, createIDXs map[string]mongo.IndexModel) error {
	names, err := getIndexNames(db.Collection(table))
	if err != nil {
		return err
	}

	for name, v := range createIDXs {
		if names[name] == true {
			continue
		}

		_, err = db.Collection(table).Indexes().CreateOne(context.Background(), v)
		if err != nil {
			return err
		}
		log.Infof("table %s create index %v", table, name)
	}

	return err
}

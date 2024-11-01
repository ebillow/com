package mgo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

type Role struct {
	ID   uint32
	Name string
}

func initMongo() {
	//err := Init("192.168.0.77:27017,192.168.0.77:27018,192.168.0.77:27019",
	//	"", "", "game", 2, "replicaSet=myRepl")
	err := Init("192.168.0.77:27017",
		"", "", "sg_center", 2, "")
	if err != nil {
		fmt.Println(err)
	}
}

func TestInit(t *testing.T) {
	initMongo()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	SyncExe(func(cli *mongo.Database) {
		data := &Role{ID: 2, Name: "test1"}
		table := cli.Collection("role")
		_, err := table.InsertOne(ctx, data)
		if err != nil {
			t.Log(err)
			t.Failed()
			return
		}
	})
}

func TestReplicaSet(t *testing.T) {
	initMongo()
	SyncExe(func(cli *mongo.Database) {
		cli.Client().UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
			if err := sessionContext.StartTransaction(); err != nil {
				return err
			}

			table := cli.Collection("role")
			data := &Role{ID: 1, Name: "test1"}
			_, err := table.InsertOne(sessionContext, data)
			if err != nil {
				_ = sessionContext.AbortTransaction(context.Background())
				return err
			}

			data.ID = 2
			_, err = table.InsertOne(sessionContext, data)
			if err != nil {
				_ = sessionContext.AbortTransaction(context.Background())
				return err
			}

			return sessionContext.CommitTransaction(context.Background())
		})
	})
}

func TestCreateIndex(t *testing.T) {
	initMongo()
	SyncExe(func(cli *mongo.Database) {
		idx := make(map[string]mongo.IndexModel)
		idx["act_act_1"] = mongo.IndexModel{
			Keys:    bson.D{{"key", 1}},
			Options: options.Index().SetUnique(false),
		}
		idx["act_act_subkey_1"] = mongo.IndexModel{
			Keys:    bson.D{{Key: "key", Value: 1}, {Key: "subkey", Value: 1}},
			Options: options.Index().SetUnique(true),
		}
		err := CreateIndex(cli, "act_act", idx)
		if err != nil {
			t.Logf("create act index err :%v", err)
			return
		}
	})
}

func TestCreateIndex2(t *testing.T) {
	initMongo()
	SyncExe(func(cli *mongo.Database) {
		idx := make(map[string]mongo.IndexModel)
		idx["test_1"] = mongo.IndexModel{
			Keys:    bson.D{{"key", 1}},
			Options: options.Index().SetUnique(true),
		}
		idx["test_test_1"] = mongo.IndexModel{
			Keys:    bson.D{{"key", 1}, {"id", 2}},
			Options: options.Index().SetUnique(true),
		}
		err := CreateIndex(cli, "test", idx)
		if err != nil {
			t.Log(err)
			t.Failed()
		}
	})
}

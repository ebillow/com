package util

import "go.mongodb.org/mongo-driver/bson"

func BsonMarshal(properties map[string]interface{}) []byte {
	data, _ := bson.Marshal(properties)
	return data
}

var reasonNameList = []string{"reason_1", "reason_2", "reason_3",
	"reason_4", "reason_5", "reason_6", "reason_7",
	"reason_8", "reason_9", "reason_10"}

func BsonMarshalWithReason(reason []interface{}, properties map[string]interface{}) []byte {
	if len(reason) > 10 {
		reason = reason[:10]
	}
	for k, v := range reason {
		reasonName := reasonNameList[k]
		properties[reasonName] = ToString(v)
	}
	data, _ := bson.Marshal(properties)
	return data
}

func BsonUnmarshal(data []byte) map[string]interface{} {
	ret := map[string]interface{}{}
	_ = bson.Unmarshal(data, ret)
	return ret
}

func GetReasonOne(reason []interface{}, pos int) string {
	if pos >= 1 && pos <= len(reason) { //[1,n]
		data := reason[pos-1]
		return ToString(data)
	}
	return ""
}

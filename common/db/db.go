package db

import (
	"github.com/lvfeiyang/proxy/common/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
)

const dbName = "leon-db"

// const mongoUrl = "mongodb://xm:784826@10.0.75.1:27017"
var mongoUrl string

func Init() {
	mongoUrl = config.ConfigVal.MongoUrl
}

func Create(cname string, data interface{}) error {
	session, err := mgo.Dial(mongoUrl) //"192.168.109.128")
	if err != nil {
		// flog.LogFile.Fatal(err)
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	err = c.Insert(data)
	if err != nil {
		return err
	}
	return nil
}

func FindOne(cname string, bm bson.M, data interface{}) error {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	if err := c.Find(bm).One(data); err != nil {
		return err
	}
	return nil
}
func FindOneById(cname string, id bson.ObjectId, data interface{}) error {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	if err := c.FindId(id).One(data); err != nil {
		return err
	}
	return nil
}
func UpdateOne(cname string, id bson.ObjectId, data interface{}) error {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	if err := c.UpdateId(id, data); err != nil {
		return err
	}
	return nil
}

func FindMany(cname string, bm bson.M, data interface{}, sort string) error {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	q := c.Find(bm)
	if "" != sort {
		q = q.Sort(sort)
	}
	if err := q.All(data); err != nil {
		return err
	}
	return nil
}
func DeleteOne(cname string, id bson.ObjectId) error {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	if err := c.RemoveId(id); err != nil {
		return err
	}
	return nil
}
func DeleteMany(cname string, bm bson.M) error {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	if _, err := c.RemoveAll(bm); err != nil {
		return err
	}
	return nil
}
func Aggregate(cname string, bm []bson.M, data interface{}) error {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(dbName).C(cname)
	if err := c.Pipe(bm).All(data); err != nil {
		return err
	}
	return nil
}
func ToMap(d interface{}) bson.M {
	out := bson.M{}
	val := reflect.ValueOf(d)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		f := val.Field(i)
		fv := f.Interface()
		if fn := typ.Field(i).Name; fn != "Id" && fv != "" {
			if "!Del" == fv {
				fv = ""
			}
			out[strings.ToLower(fn)] = fv
		}
	}
	return out
}

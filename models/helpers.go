package models

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ModelSettings struct {
	Path, Noun, CollectionName string
	DataStruct                 interface{}
}

type FindAllOutputStruct struct {
	links struct {
		self, prev, next string
	}

	meta struct {
		page struct {
			offset, limit, total int
		}
	}

	data interface{}
}

func FindAll(rootUrl string, collection *mgo.Collection, query bson.M, pageOffset, pageLimit int) (bson.M, error) {

	pageTotal, err := collection.Find(bson.M{}).Count()
	if err != nil {
		return nil, err
	}

	pageLinks := bson.M{"self": fmt.Sprintf("%s?page[offset]=%d&page[limit]=%d", rootUrl, pageOffset, pageLimit)}
	pageOffsetPrev := pageOffset - pageLimit
	if pageOffsetPrev >= 0 {
		pageLinks["prev"] = fmt.Sprintf("%s?page[offset]=%d&page[limit]=%d", rootUrl, pageOffsetPrev, pageLimit)
	}

	pageOffsetNext := pageOffset + pageLimit
	if pageOffsetNext < pageTotal {
		pageLinks["next"] = fmt.Sprintf("%s?page[offset]=%d&page[limit]=%d", rootUrl, pageOffsetNext, pageLimit)
	}

	query["deleted_at"] = bson.M{"$exists": false}

	usr := &[]bson.M{}
	if err := collection.Find(query).Skip(pageOffset).Limit(pageLimit).Sort("Name").All(usr); err != nil {
		return nil, err
	}

	data := (bson.M{
		"links": pageLinks,
		"meta":  bson.M{"page": bson.M{"offset": pageOffset, "limit": pageLimit, "total": pageTotal}},
		"data":  usr})

	return data, err
}

func IsExists(collection *mgo.Collection, query *bson.M) bool {
	usr := []bson.M{}
	if err := collection.Find(query).Limit(1).All(&usr); err != nil {
		return false
	}
	return len(usr) != 0
}

func FindOne(collection *mgo.Collection, query *bson.M) (bson.M, error) {
	usr := bson.M{}
	if err := collection.Find(query).One(&usr); err != nil {
		return nil, err
	}
	return usr, nil
}

func FindId(collection *mgo.Collection, id string) (*bson.M, error) {
	usr := &bson.M{}

	// query["deleted_at"] = bson.M{"$exists": false}
	if err := collection.Find(bson.M{"_id": bson.ObjectIdHex(id), "deleted_at": bson.M{"$exists": false}}).One(usr); err != nil {
		return nil, err
	}
	return usr, nil
}

func Update(collection *mgo.Collection, id string, usr *bson.M) error {

	fmt.Println("Update data for model", id, usr)

	if err := collection.UpdateId(bson.ObjectIdHex(id), &bson.M{"$set": usr}); err != nil {
		fmt.Println("Can't update in model", err)
		return err
	}

	return nil
}

func Create(collection *mgo.Collection, usr *bson.M) error {
	if err := collection.Insert(usr); err != nil {
		return err
	}
	// on future put also the id after insert
	return nil
}

func Remove(collection *mgo.Collection, id string) error {
	/*
		if err := collection.RemoveId(bson.ObjectIdHex(id)); err != nil {
			return err
		}
	*/
	if err := collection.UpdateId(bson.ObjectIdHex(id), &bson.M{"$set": bson.M{"deleted_at": time.Now()}}); err != nil {
		fmt.Println("Can't update in model", err)
		return err
	}
	return nil
}

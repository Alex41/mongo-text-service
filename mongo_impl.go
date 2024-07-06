package translation_service

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type client[ID comparable, V, A any] struct {
	coll *mongo.Collection
}

var upsert = options.Update().SetUpsert(true)

type translation[ID comparable, V, A any] struct {
	Key       ID             `bson:"_id"`
	Value     map[Language]V `bson:"trs"`
	AddFields A              `bson:"adt"`
}

func (t translation[ID, V, A]) GetID() ID {
	return t.Key
}

func (t translation[ID, V, A]) GetTranslates() map[Language]V {
	return t.Value
}

func (t translation[ID, V, A]) GetAdditional() A {
	return t.AddFields
}

func (m *client[ID, V, A]) GetTranslations(c context.Context, key ID) (Response[ID, V, A], error) {
	var (
		filter = bson.M{"_id": key}
		result translation[ID, V, A]
	)

	err := m.coll.FindOne(c, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = TextNotFound
	}

	return result, err
}

func (m *client[ID, V, A]) GetAllTranslations(c context.Context, filter bson.M) ([]Response[ID, V, A], error) {
	crs, err := m.coll.Find(c, filter)
	if err != nil {
		return nil, err
	}

	var dnl []translation[ID, V, A]
	err = crs.All(c, &dnl)

	resp := make([]Response[ID, V, A], len(dnl))
	for i, re := range dnl {
		resp[i] = re
	}

	return resp, err
}

func (m *client[ID, V, A]) UpsertAll(c context.Context, key ID, mp map[Language]V, a A) error {
	var set = bson.M{"adt": a}

	for lang, v := range mp {
		set[fmt.Sprintf("trs.%s", lang)] = v
	}

	filter := bson.M{"_id": key}
	update := bson.M{"$set": set}

	_, err := m.coll.UpdateOne(c, filter, update, upsert)

	return err
}

func (m *client[ID, V, A]) Upsert(c context.Context, key ID, lang Language, v V) error {
	filter := bson.M{"_id": key}
	update := bson.M{"$set": bson.M{
		fmt.Sprintf("trs.%s", lang): v,
	}}

	_, err := m.coll.UpdateOne(c, filter, update, upsert)

	return err
}

func (m *client[ID, V, A]) SetAdditional(c context.Context, key ID, a A) error {
	filter := bson.M{"_id": key}
	update := bson.M{"$set": bson.M{"adt": a}}

	_, err := m.coll.UpdateOne(c, filter, update)

	return err
}

func (m *client[ID, V, A]) Delete(c context.Context, key ID) error {
	_, err := m.coll.DeleteOne(c, bson.M{"_id": key})
	return err
}

func (m *client[ID, V, A]) GetTranslation(c context.Context, key ID, lang Language) (*V, error) {
	var (
		filter = bson.M{
			"_id":                       key,
			fmt.Sprintf("trs.%s", lang): bson.M{"$exists": true},
		}
		result translation[ID, V, A]
	)

	err := m.coll.FindOne(c, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, TextNotFound
	}

	// todo: use projection
	value, exists := result.Value[lang]
	if !exists {
		return nil, TextNotFound
	}

	return &value, err
}

//goland:noinspection GoUnusedExportedFunction
func NewClient[ID comparable, V, A any](coll *mongo.Collection) TranslationService[ID, V, A] {
	return &client[ID, V, A]{coll: coll}
}

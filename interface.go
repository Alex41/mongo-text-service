package translation_service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type TranslationService[ID comparable, V, A any] interface {
	GetAllTranslations(_ context.Context, filter bson.M) ([]Response[ID, V, A], error)
	GetTranslation(context.Context, ID) (Response[ID, V, A], error)
	SetAdditional(context.Context, ID, A) error
	Upsert(context.Context, ID, Language, V) error
	UpsertAll(context.Context, ID, map[Language]V, A) error
	Delete(context.Context, ID) error
}

type Response[ID comparable, V, A any] interface {
	GetTranslates() map[Language]V
	GetAdditional() A
	GetID() ID
}

type Language string

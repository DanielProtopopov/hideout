package main

import (
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	"hideout/api"
	apiconfig "hideout/cmd/api/config"
	"hideout/internal/common/model"
	"hideout/internal/paths"
	secrets2 "hideout/internal/secrets"
	"hideout/services/secrets"
	"hideout/structs"
	"log"
	"runtime/debug"
	"time"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("[ERROR] Error loading %s file", ".env")
	}

	apiconfig.Init(ctx)
	log.Println("Configuration was successfully loaded")

	go func() {
		t := time.Tick(time.Minute)
		for {
			<-t
			debug.FreeOSMemory()
		}
	}()

	secretsSvc, errCreateService := secrets.NewService(ctx, apiconfig.Settings.Repository, &structs.Paths, &structs.Secrets)
	if errCreateService != nil {
		log.Fatal(errCreateService)
	}

	if apiconfig.Settings.Repository.PreloadInMemory {
		errReload := secretsSvc.Load(ctx)
		if errReload != nil {
			log.Fatal(errReload)
		}
	}

	rootPath, _ := secretsSvc.CreatePath(ctx, paths.Path{Model: model.Model{ID: 0}, Name: ""})
	testPath, _ := secretsSvc.CreatePath(ctx, paths.Path{ParentID: rootPath.ID, Name: "test"})

	anotherTestPath, _ := secretsSvc.CreatePath(ctx, paths.Path{ParentID: rootPath.ID, Name: "another-test"})
	yetAnotherTestPath, _ := secretsSvc.CreatePath(ctx, paths.Path{ParentID: rootPath.ID, Name: "yet-another-test"})

	rootSecret, _ := secretsSvc.CreateSecret(ctx, secrets2.Secret{PathID: rootPath.ID, Name: "Root secret", Value: "123", Type: "integer"})
	_, _ = secretsSvc.CreateSecret(ctx, secrets2.Secret{PathID: testPath.ID, Name: "Secret #1", Value: "123", Type: "integer"})
	_, _ = secretsSvc.CreateSecret(ctx, secrets2.Secret{PathID: testPath.ID, Name: "Secret #2", Value: "456", Type: "integer"})
	_, _ = secretsSvc.CreateSecret(ctx, secrets2.Secret{PathID: testPath.ID, Name: "Secret #3", Value: "789", Type: "integer"})

	tree, errGetTree := secretsSvc.Tree(ctx, rootPath.ID)
	if errGetTree != nil {
		log.Fatal(errGetTree)
	}

	jsonResult, errMarshal := json.Marshal(tree)
	if errMarshal != nil {
		log.Fatal(errMarshal)
	}
	log.Println(string(jsonResult))

	_, _, errCopy := secretsSvc.Copy(ctx, []*paths.Path{testPath}, []*secrets2.Secret{rootSecret},
		rootPath.ID, anotherTestPath.ID)
	if errCopy != nil {
		log.Fatal(errCopy)
	}

	tree, errGetTree = secretsSvc.Tree(ctx, rootPath.ID)
	if errGetTree != nil {
		log.Fatal(errGetTree)
	}

	jsonResult, errMarshal = json.Marshal(tree)
	if errMarshal != nil {
		log.Fatal(errMarshal)
	}
	log.Println(string(jsonResult))

	_, _, errCopy = secretsSvc.Copy(ctx, []*paths.Path{anotherTestPath, testPath}, []*secrets2.Secret{rootSecret},
		rootPath.ID, yetAnotherTestPath.ID)
	if errCopy != nil {
		log.Fatal(errCopy)
	}

	tree, errGetTree = secretsSvc.Tree(ctx, rootPath.ID)
	if errGetTree != nil {
		log.Fatal(errGetTree)
	}

	jsonResult, errMarshal = json.Marshal(tree)
	if errMarshal != nil {
		log.Fatal(errMarshal)
	}
	log.Println(string(jsonResult))

	_, _, errDelete := secretsSvc.Delete(ctx, []*paths.Path{anotherTestPath}, nil, rootPath.ID, false)
	if errDelete != nil {
		log.Fatal(errDelete)
	}

	tree, errGetTree = secretsSvc.Tree(ctx, rootPath.ID)
	if errGetTree != nil {
		log.Fatal(errGetTree)
	}

	jsonResult, errMarshal = json.Marshal(tree)
	if errMarshal != nil {
		log.Fatal(errMarshal)
	}
	log.Println(string(jsonResult))

	api.Serve()
}

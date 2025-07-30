package main

import (
	"context"
	"encoding/json"
	"hideout/api"
	apiconfig "hideout/cmd/api/config"
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

	apiconfig.Init(ctx)
	log.Println("Configuration was successfully loaded")

	go func() {
		t := time.Tick(time.Minute)
		for {
			<-t
			debug.FreeOSMemory()
		}
	}()

	secretsSvc, errCreateService := secrets.NewService(ctx, secrets.Config{}, &structs.Paths, &structs.Secrets,
		secrets.RepositoryType_Redis, true)
	if errCreateService != nil {
		log.Fatal(errCreateService)
	}

	rootPath, _ := secretsSvc.CreatePath(ctx, 0, "")
	testPath, _ := secretsSvc.CreatePath(ctx, rootPath.ID, "test")
	anotherTestPath, _ := secretsSvc.CreatePath(ctx, rootPath.ID, "another-test")
	yetAnotherTestPath, _ := secretsSvc.CreatePath(ctx, rootPath.ID, "yet-another-test")

	rootSecret, _ := secretsSvc.CreateSecret(ctx, rootPath.ID, "Root secret", "123", "Integer")
	_, _ = secretsSvc.CreateSecret(ctx, testPath.ID, "Secret #1", "123", "Integer")
	_, _ = secretsSvc.CreateSecret(ctx, testPath.ID, "Secret #2", "456", "Integer")
	_, _ = secretsSvc.CreateSecret(ctx, testPath.ID, "Secret #3", "789", "Integer")

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

	_, _, errDelete := secretsSvc.Delete(ctx, []*paths.Path{anotherTestPath}, nil, rootPath.ID)
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

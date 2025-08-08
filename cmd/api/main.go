package main

import (
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	"hideout/api"
	apiconfig "hideout/cmd/api/config"
	"hideout/internal/folders"
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

	secretsSvc, errCreateService := secrets.NewService(ctx, apiconfig.Settings.SecretsRepository, apiconfig.Settings.FoldersRepository, &structs.Folders, &structs.Secrets)
	if errCreateService != nil {
		log.Fatal(errCreateService)
	}

	rootFolder, _ := secretsSvc.CreateFolder(ctx, folders.Folder{Name: ""})
	testFolder, _ := secretsSvc.CreateFolder(ctx, folders.Folder{ParentID: rootFolder.ID, Name: "test"})

	anotherTestFolder, _ := secretsSvc.CreateFolder(ctx, folders.Folder{ParentID: rootFolder.ID, Name: "another-test"})
	yetAnotherTestFolder, _ := secretsSvc.CreateFolder(ctx, folders.Folder{ParentID: rootFolder.ID, Name: "yet-another-test"})

	rootSecret, _ := secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: rootFolder.ID, Name: "Root secret", Value: "123", Type: "integer"})
	_, _ = secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: testFolder.ID, Name: "Secret #1", Value: "123", Type: "integer"})
	_, _ = secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: testFolder.ID, Name: "Secret #2", Value: "456", Type: "integer"})
	_, _ = secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: testFolder.ID, Name: "Secret #3", Value: "789", Type: "integer"})

	tree, errGetTree := secretsSvc.Tree(ctx, rootFolder.ID)
	if errGetTree != nil {
		log.Fatal(errGetTree)
	}

	jsonResult, errMarshal := json.Marshal(tree)
	if errMarshal != nil {
		log.Fatal(errMarshal)
	}
	log.Println(string(jsonResult))

	_, _, errCopy := secretsSvc.Copy(ctx, []*folders.Folder{testFolder}, []*secrets2.Secret{rootSecret},
		rootFolder.ID, anotherTestFolder.ID)
	if errCopy != nil {
		log.Fatal(errCopy)
	}

	tree, errGetTree = secretsSvc.Tree(ctx, rootFolder.ID)
	if errGetTree != nil {
		log.Fatal(errGetTree)
	}

	jsonResult, errMarshal = json.Marshal(tree)
	if errMarshal != nil {
		log.Fatal(errMarshal)
	}
	log.Println(string(jsonResult))

	_, _, errCopy = secretsSvc.Copy(ctx, []*folders.Folder{anotherTestFolder, testFolder}, []*secrets2.Secret{rootSecret},
		rootFolder.ID, yetAnotherTestFolder.ID)
	if errCopy != nil {
		log.Fatal(errCopy)
	}

	tree, errGetTree = secretsSvc.Tree(ctx, rootFolder.ID)
	if errGetTree != nil {
		log.Fatal(errGetTree)
	}

	jsonResult, errMarshal = json.Marshal(tree)
	if errMarshal != nil {
		log.Fatal(errMarshal)
	}
	log.Println(string(jsonResult))

	_, _, errDelete := secretsSvc.Delete(ctx, []*folders.Folder{anotherTestFolder}, nil, rootFolder.ID, false)
	if errDelete != nil {
		log.Fatal(errDelete)
	}

	tree, errGetTree = secretsSvc.Tree(ctx, rootFolder.ID)
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

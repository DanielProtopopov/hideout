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

	rootFolder, errCreateRootFolder := secretsSvc.CreateFolder(ctx, folders.Folder{Name: ""})
	if errCreateRootFolder != nil {
		log.Fatal(errCreateRootFolder)
	}
	testFolder, errCreateTestFolder := secretsSvc.CreateFolder(ctx, folders.Folder{ParentID: rootFolder.ID, Name: "test"})
	if errCreateTestFolder != nil {
		log.Fatal(errCreateTestFolder)
	}
	anotherTestFolder, errCreateAnotherTestFolder := secretsSvc.CreateFolder(ctx, folders.Folder{ParentID: rootFolder.ID, Name: "another-test"})
	if errCreateAnotherTestFolder != nil {
		log.Fatal(errCreateAnotherTestFolder)
	}
	yetAnotherTestFolder, errCreateYetAnotherTestFolder := secretsSvc.CreateFolder(ctx, folders.Folder{ParentID: rootFolder.ID, Name: "yet-another-test"})
	if errCreateYetAnotherTestFolder != nil {
		log.Fatal(errCreateYetAnotherTestFolder)
	}
	rootSecret, errCreateRootSecret := secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: rootFolder.ID, Name: "Root secret", Value: "123", Type: "integer"})
	if errCreateRootSecret != nil {
		log.Fatal(errCreateRootSecret)
	}
	_, errCreateSecret1 := secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: testFolder.ID, Name: "Secret #1", Value: "123", Type: "integer"})
	if errCreateSecret1 != nil {
		log.Fatal(errCreateSecret1)
	}
	_, errCreateSecret2 := secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: testFolder.ID, Name: "Secret #2", Value: "456", Type: "integer"})
	if errCreateSecret2 != nil {
		log.Fatal(errCreateSecret2)
	}
	_, errCreateSecret3 := secretsSvc.CreateSecret(ctx, secrets2.Secret{FolderID: testFolder.ID, Name: "Secret #3", Value: "789", Type: "integer"})
	if errCreateSecret3 != nil {
		log.Fatal(errCreateSecret3)
	}

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
